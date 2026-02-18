package luna

import (
	"html/template"
	"strings"
)

type htmlWidget struct {
	widgetBase `yaml:",inline"`
	Source     template.HTML `yaml:"source"`
}

func (widget *htmlWidget) initialize() error {
	widget.withTitle("").withError(nil)

	return nil
}

func (widget *htmlWidget) Render() template.HTML {
	result := string(widget.Source)
	if widget.Notifications && notificationsEnabledForWidget(widget.Type) && shouldUseGenericNotifications(widget.Type) {
		if widget.lastRenderedHTML != "" && widget.lastRenderedHTML != result {
			displayTitle := widget.Title
			if strings.TrimSpace(displayTitle) == "" {
				displayTitle = widget.Type
			}
			body := "Widget content changed."
			if strings.TrimSpace(widget.TitleURL) != "" {
				body = body + "\nURL: " + widget.TitleURL
			}
			sendWidgetNotification(widget.Type, "Widget: "+displayTitle, body, "info")
		}
		widget.lastRenderedHTML = result
	}

	return widget.Source
}
