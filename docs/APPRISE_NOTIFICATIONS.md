# Apprise Notifications Implementation for Luna

## Overview
Integrated Apprise API notification support for all widget types in Luna with flexible control via environment variables and per-widget/per-site configuration.

## Features

### 1. Environment Configuration (.env)
Load environment variables from a `.env` file at startup (no restart needed for config changes, but restart required for env changes).

```env
# Your Apprise API endpoint
APPRISE_URL=http://apprise-api:8000/notify/

# Notification providers (comma-separated)
# Supports: Telegram, Discord, Gotify, SMTP, Slack, etc.
APPRISE_PROVIDERS=tgram://bottoken/ChatID,discord://webhookid/token

# Master toggle (enables all widgets if no specific override)
NOTIFY_ALL=0

# Per-widget filters (overrides NOTIFY_ALL)
NOTIFY_MONITOR=1
NOTIFY_RSS=0
NOTIFY_REDDIT=0
NOTIFY_YOUTUBE=1
NOTIFY_CUSTOM_API=0
NOTIFY_DOCKER_CONTAINERS=0
# ... other widgets follow NOTIFY_<WIDGET_TYPE> pattern (hyphens → underscores)
```

### 2. Luna Configuration (luna.yml)
Enable notifications at widget or monitor-site level:

```yaml
- type: monitor
  title: Services
  sites:
    - name: Proxmox Node
      url: https://192.168.1.10
      notifications: true        # Per-site notification opt-in
    - name: Jellyfin
      url: https://jellyfin.mydomain.com
      notifications: false       # Opt-out for this site

- type: rss
  title: News
  notifications: true            # Widget-level opt-in
  feeds:
    - url: https://example.com/rss

- type: docker-containers
  title: Containers
  notifications: true
```

### 3. Notification Behavior

#### All Widgets (Generic)
Any change in rendered content (HTML) triggers a notification:
- Message: `"Widget: <title>" "Widget content changed. URL: <title-url>"`
- Applies to: calendar, weather, clock, bookmarks, extension, etc.
- Only notifies on actual changes, not on first load

#### Monitor
Status change per site, with detailed info:
- Message: `"Monitor: <site-title>"` `"Status changed from <old> to <new> for <url> (status <code>)"`
- Includes error messages if request failed
- Per-site override: `notifications: true/false`

#### RSS Feeds
Feed update detection (new items or removed items):
- Message: `"RSS: <feed-title>"` 
- Lists up to 3 new items with URLs
- Detects any change in feed items, not just new entries

#### Reddit
Subreddit update detection:
- Message: `"Reddit: <subreddit>"`
- Lists up to 3 new posts with URLs
- Detects any change in available posts

#### YouTube
Channel/playlist update detection:
- Message: `"YouTube: <title>"`
- Lists up to 3 new videos with URLs
- Detects any change in available videos

#### Custom API
Data change notifications:
- Message: `"Custom API: <title>"` `"Widget data changed. URL: <api-url>"`
- Triggered whenever the rendered output changes

### 4. Control Flow

1. **Enable Notifications:**
   - Set `NOTIFY_ALL=1` in `.env` to enable all widgets globally, OR
   - Set `NOTIFY_<WIDGET_TYPE>=1` for individual widgets
   - Per-widget `notifications: true` in config enables for that widget/site

2. **Disable Notifications:**
   - Set `NOTIFY_ALL=0` and leave widget-specific vars unset (default behavior)
   - Set specific `NOTIFY_<WIDGET_TYPE>=0` to override `NOTIFY_ALL=1`
   - Set `notifications: false` in config to disable for specific widget/site

3. **Override Hierarchy:**
   ```
   Per-widget config (notifications: true/false)
          ↓
   Widget-specific env (NOTIFY_RSS=1, NOTIFY_MONITOR=1)
          ↓
   Master toggle (NOTIFY_ALL=1)
          ↓
   Default: disabled
   ```

### 5. Implementation Details

**New Files:**
- `internal/luna/env.go` - .env file parser
- `internal/luna/notifications.go` - Apprise API client & controls

**Modified Files:**
- `internal/luna/main.go` - Load .env on startup
- `internal/luna/widget.go` - Generic notification tracking & sending
- `internal/luna/widget-monitor.go` - Monitor status change notifications
- `internal/luna/widget-rss.go` - Feed update notifications
- `internal/luna/widget-reddit.go` - Post update notifications
- `internal/luna/widget-videos.go` - Video update notifications
- `internal/luna/widget-custom-api.go` - Data change notifications
- `internal/luna/widget-html.go` - Generic HTML widget notifications
- `internal/luna/widget-shared.go` - Added post ID field for tracking
- `docs/configuration.md` - Configuration documentation

### 6. Docker Compose Example

```yaml
services:
  apprise:
    image: linuxserver/apprise:latest
    container_name: apprise-api
    ports:
      - "8000:8000"
    environment:
      TZ: UTC

  luna:
    image: ghcr.io/luna-page/luna:main
    container_name: luna
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./config:/app/config:ro
      - ./.env:/app/.env:ro
    environment:
      - TZ=UTC
      - luna_CONFIG=/app/config/luna.yml
    depends_on:
      - apprise
```

### 7. Notes

- **Env vars loaded at startup:** Changes to `.env` require Luna restart
- **No duplicate notifications:** Widgets with custom logic (monitor, RSS, etc.) don't trigger generic notifications
- **First load safe:** Initial widget state doesn't trigger notifications
- **Rate limiting:** None (Apprise handles rate limiting per service)
- **Async sending:** Notifications sent asynchronously to avoid blocking widget updates
- **Failure handling:** Notification failures log but don't fail widget updates

## Testing Workflow

1. Configure `.env`:
   ```env
   APPRISE_URL=http://localhost:8000/notify/
   APPRISE_PROVIDERS=json://localhost:9999
   NOTIFY_ALL=1
   ```

2. Enable in config:
   ```yaml
   - type: monitor
     sites:
       - title: Test
         url: https://httpbin.org/status/200
         notifications: true
   ```

3. Trigger a change (e.g., change URL to an unreachable one)

4. Check Apprise logs for received notifications

5. Adjust `NOTIFY_MONITOR=0` to disable, or `notifications: false` per-site to disable
