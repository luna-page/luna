package luna

import (
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func init() {
	registerMarketProvider("stooq", &stooqProvider{}, nil) // no rate limit needed
}

type stooqProvider struct{}

func (p *stooqProvider) FetchMarkets(marketRequests []marketRequest, _ string) (marketList, error) {
	// Step 1: Batch quote for current prices
	symbols := make([]string, len(marketRequests))
	symbolMap := make(map[string]marketRequest, len(marketRequests))
	for i, r := range marketRequests {
		symbols[i] = r.Symbol
		symbolMap[r.Symbol] = r
	}

	quoteURL := fmt.Sprintf("https://stooq.com/q/l/?s=%s&f=sd2t2ohlcv&h&e=csv",
		strings.Join(symbols, "+"))

	quoteResp, err := defaultHTTPClient.Do(mustNewRequest("GET", quoteURL))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errNoContent, err)
	}
	defer quoteResp.Body.Close()

	quotes, err := parseStooqCSV(quoteResp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse stooq quotes: %v", errNoContent, err)
	}

	// Step 2: Fetch historical data for sparkline charts (one request per symbol)
	now := time.Now()
	d2 := now.Format("20060102")
	d1 := now.AddDate(0, -1, 0).Format("20060102")

	histRequests := make([]*http.Request, len(marketRequests))
	for i, r := range marketRequests {
		url := fmt.Sprintf("https://stooq.com/q/d/l/?s=%s&d1=%s&d2=%s&i=d", r.Symbol, d1, d2)
		histRequests[i], _ = http.NewRequest("GET", url, nil)
	}

	type histResult struct {
		prices []float64
		err    error
	}

	histJob := newJob(func(req *http.Request) (histResult, error) {
		resp, err := defaultHTTPClient.Do(req)
		if err != nil {
			return histResult{}, err
		}
		defer resp.Body.Close()

		prices, err := parseStooqHistoricalCSV(resp.Body)
		return histResult{prices: prices, err: err}, nil
	}, histRequests)

	histResponses, histErrs, err := workerPoolDo(histJob)
	if err != nil {
		// Continue without sparklines if historical fetch fails entirely
		slog.Warn("Failed to fetch historical data from Stooq", "error", err)
	}

	// Step 3: Combine quotes + historical into market list
	markets := make(marketList, 0, len(marketRequests))
	var failed int

	for i, req := range marketRequests {
		quote, exists := quotes[req.Symbol]
		if !exists {
			failed++
			slog.Error("No quote data from Stooq", "symbol", req.Symbol)
			continue
		}

		if quote.close == 0 {
			failed++
			slog.Error("Stooq returned no data (N/D)", "symbol", req.Symbol)
			continue
		}

		// Sparkline from historical data
		var points string
		if histErrs != nil && histErrs[i] == nil && histResponses[i].err == nil {
			prices := histResponses[i].prices
			if len(prices) > marketChartDays {
				prices = prices[len(prices)-marketChartDays:]
			}
			if len(prices) > 1 {
				points = svgPolylineCoordsFromYValues(100, 50, maybeCopySliceWithoutZeroValues(prices))
			}
		}

		// Calculate percent change from previous close
		var pctChange float64
		if quote.prevClose != 0 {
			pctChange = percentChange(quote.close, quote.prevClose)
		} else if histResponses != nil && histErrs[i] == nil && len(histResponses[i].prices) >= 2 {
			prices := histResponses[i].prices
			pctChange = percentChange(prices[len(prices)-1], prices[len(prices)-2])
		}

		// Detect currency from symbol suffix
		currency := stooqCurrencyFromSymbol(req.Symbol)

		markets = append(markets, market{
			marketRequest: req,
			Price:         quote.close,
			Currency:      currency,
			PriceHint:     2,
			Name: ternary(req.CustomName == "",
				req.Symbol,
				req.CustomName,
			),
			PercentChange:  pctChange,
			SvgChartPoints: points,
		})
	}

	return marketsFailed(len(marketRequests), failed, markets)
}

type stooqQuote struct {
	close     float64
	prevClose float64 // derived from open as approximation
}

func parseStooqCSV(r io.Reader) (map[string]stooqQuote, error) {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) <= 1 {
		return nil, fmt.Errorf("no quote data in CSV response")
	}

	quotes := make(map[string]stooqQuote, len(records)-1)

	for i, record := range records {
		if i == 0 { // skip header
			continue
		}
		if len(record) < 8 {
			continue
		}

		symbol := record[0]
		closeStr := record[6]
		openStr := record[3]

		closeVal, err := strconv.ParseFloat(closeStr, 64)
		if err != nil {
			continue // N/D values
		}
		openVal, _ := strconv.ParseFloat(openStr, 64)

		quotes[symbol] = stooqQuote{
			close:     closeVal,
			prevClose: openVal, // open approximates previous close
		}
	}

	return quotes, nil
}

func parseStooqHistoricalCSV(r io.Reader) ([]float64, error) {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) <= 1 {
		return nil, fmt.Errorf("no historical data in CSV response")
	}

	prices := make([]float64, 0, len(records)-1)

	for i, record := range records {
		if i == 0 { // skip header
			continue
		}
		if len(record) < 5 {
			continue
		}

		closeVal, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			continue
		}
		prices = append(prices, closeVal)
	}

	return prices, nil
}

func stooqCurrencyFromSymbol(symbol string) string {
	parts := strings.SplitN(symbol, ".", 2)
	if len(parts) < 2 {
		return "$"
	}

	switch strings.ToUpper(parts[1]) {
	case "US":
		return "$"
	case "UK":
		return "£"
	case "DE", "FR", "ES", "IT", "NL":
		return "€"
	case "JP":
		return "¥"
	case "HK":
		return "HK$"
	case "V": // crypto (BTC.V, ETH.V)
		return "$"
	default:
		return "$"
	}
}

func mustNewRequest(method, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	return req
}
