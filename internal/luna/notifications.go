package luna

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type notificationSettings struct {
	AppriseURL string
	Providers  string
}

var (
	notificationSettingsOnce   sync.Once
	notificationSettingsCached notificationSettings
)

func getNotificationSettings() notificationSettings {
	notificationSettingsOnce.Do(func() {
		notificationSettingsCached = notificationSettings{
			AppriseURL: strings.TrimSpace(os.Getenv("APPRISE_URL")),
			Providers:  strings.TrimSpace(os.Getenv("APPRISE_PROVIDERS")),
		}
	})

	return notificationSettingsCached
}

func NotificationsEnabledForWidget(widgetType string) bool {
	envVar := notificationEnvVarForWidget(widgetType)
	if envVar == "" {
		if value, found := os.LookupEnv("NOTIFY_ALL"); found {
			return parseEnvBool(value)
		}
		return false
	}

	value, found := os.LookupEnv(envVar)
	if !found {
		if value, found := os.LookupEnv("NOTIFY_ALL"); found {
			return parseEnvBool(value)
		}
		return false
	}

	return parseEnvBool(value)
}

func parseEnvBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return false
	}
}

func notificationEnvVarForWidget(widgetType string) string {
	switch widgetType {
	case "videos":
		return "NOTIFY_YOUTUBE"
	case "monitor":
		return "NOTIFY_MONITOR"
	case "rss":
		return "NOTIFY_RSS"
	case "reddit":
		return "NOTIFY_REDDIT"
	case "custom-api":
		return "NOTIFY_CUSTOM_API"
	default:
		if widgetType == "" {
			return ""
		}
		return "NOTIFY_" + strings.ToUpper(strings.ReplaceAll(widgetType, "-", "_"))
	}
}

func ShouldUseGenericNotifications(widgetType string) bool {
	switch widgetType {
	case "monitor", "rss", "reddit", "videos", "custom-api":
		return false
	default:
		return true
	}
}

func StringSetChanged(previous, current map[string]struct{}) bool {
	if len(previous) != len(current) {
		return true
	}
	for key := range current {
		if _, exists := previous[key]; !exists {
			return true
		}
	}
	return false
}

func SendWidgetNotification(widgetType string, title string, body string, notifyType string) {
	if !NotificationsEnabledForWidget(widgetType) {
		return
	}

	settings := getNotificationSettings()
	if settings.AppriseURL == "" || settings.Providers == "" {
		return
	}

	go func() {
		if err := SendAppriseNotification(settings, title, body, notifyType); err != nil {
			log.Printf("failed to send apprise notification: %v", err)
		}
	}()
}

func SendAppriseNotification(settings notificationSettings, title string, body string, notifyType string) error {
	payload := map[string]any{
		"title": title,
		"body":  body,
		"urls":  settings.Providers,
	}

	if notifyType != "" {
		payload["type"] = notifyType
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := strings.TrimRight(settings.AppriseURL, "/")
	if !strings.HasSuffix(url, "/notify") && !strings.HasSuffix(url, "/notify/") {
		url = url + "/notify"
	}
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}

	request, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("apprise returned status %d", response.StatusCode)
	}

	return nil
}
