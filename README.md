<p align="center"><img src="docs/logo.png"></p>
<h1 align="center">Luna </h1> <br/ >
<p align="center"> • fork from Glance •</p>
<p align="center">
  <a href="#installation">Install</a> •
  <a href="docs/configuration.md#configuring-luna">Configuration</a> •

</p>
<p align="center">
  <a href="/docs/LIVE_EVENTS_IMPLEMENTATION.md">New : Live Events Implementation</a> •
  <a href="https://github.com/luna-page/community-widgets">Community widgets</a> •
  <a href="docs/preconfigured-pages.md">Preconfigured pages</a> •
  <a href="docs/themes.md">Themes</a>
</p>

<p align="center">Luna is a lightweight, open-source dashboard for homelabs and infrastructure.<br> simple. fast. open.</p>

![](docs/images/readme-main-image.png)

## Features
### Various widgets
* RSS feeds
* Subreddit posts
* Hacker News posts
* Weather forecasts
* YouTube channel uploads
* Twitch channels
* Market prices
* Docker containers status
* Server stats
* Custom widgets
* [and many more...](docs/configuration.md#configuring-luna)

### Fast and lightweight
* Low memory usage
* Few dependencies
* Minimal vanilla JS
* Single <20mb binary available for multiple OSs & architectures and just as small Docker container
* Uncached pages usually load within ~1s (depending on internet speed and number of widgets)

### Tons of customizability
* Different layouts
* As many pages/tabs as you need
* Numerous configuration options for each widget
* Multiple styles for some widgets
* Custom CSS

### Optimized for mobile devices
Because you'll want to take it with you on the go.

![](docs/images/mobile-preview.png)

### Themeable
Easily create your own theme by tweaking a few numbers or choose from one of the [already available themes](docs/themes.md).

## Installation
### Docker Install
Create a new directory called `config` and add [`luna.yml`](docs/luna.yml) file in the directory
```bash
mkdir config && wget -O config/luna.yml https://raw.githubusercontent.com/luna-page/luna/refs/heads/main/docs/luna.yml
```

```yaml
services:
  luna:
    image: ghcr.io/luna-page/luna:main
    container_name: luna
    restart: unless-stopped
    # If you need luna to see services running directly on the host (e.g. DNS on port 53)
      # you can use network_mode: host or add-host
    ports:
      - "8080:8080"
    volumes:
      - ./config:/app/config:ro # add luna.yml to config folder
      # If you have custom assets (CSS/JS) that you want to test without rebuilding
      # - ./public:/app/public:ro
    environment:
      - TZ=Etc/UTC
      - luna_CONFIG=/app/luna.yml
    # Important for SSE (Live Events) data flow
    logging:
      driver: "json-file"
      options:
        max-size: "10mb"
        max-file: "3"
```
```bash
docker compose up -d
```
If you encounter any issues, you can check the logs by running:

```bash
docker logs luna
```

### Update Luna
```bash
docker compose down
docker compose pull
docker compose up -d
```
### Luna Installation on Linux environment (LXC, VPS, or Bare Metal).
run:
```bash
bash -c "$(curl -fsSL https://raw.githubusercontent.com/luna-page/luna/main/ct/luna.sh)"
```
** To update the lates build just type update.

