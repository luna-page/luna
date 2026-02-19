<p align="center"><img src="docs/logo.png"></p>
<h1 align="center">Luna </h1> <br/ >
<p align="center"> • fork from Glance •</p>
<p align="center">
  <a href="#installation">Install</a> •
  <a href="docs/configuration.md#configuring-luna">Configuration</a> •

</p>
<p align="center">
  <a href="/docs/LIVE_EVENTS_IMPLEMENTATION.md">New : Live Events Implementation</a> •
  <a href="/docs/APPRISE_NOTIFICATIONS.md">New : Notification</a> •
  <a href="https://github.com/luna-page/agent">Luna Agent</a> •
  <a href="https://github.com/luna-page/community-widgets">Community widgets</a> •
  <a href="docs/preconfigured-pages.md">Preconfigured pages</a> •
  <a href="docs/themes.md">Themes</a>
</p>

<p align="center">Luna is a lightweight, open-source dashboard for homelabs and infrastructure.<br> simple. fast. open.</p>

![](docs/images/readme-main-image.png)

## Features
### What's new in Luna
* Real-time services, monitor status change
* Notification  [read more]( docs/APPRISE_NOTIFICATIONS.md)
* Luna Agent has now support for Synology & Unraid [read more](https://github.com/luna-page/agent)
* more to come..
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
Choose one of the following methods:

<details>
<summary><strong>Docker container without Apprise Notification (if you already have Apprise)</strong></summary>


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
    # If you need luna to see services running directly on the host (e.g. DNS on port 53) and to proper use the notification system
    network_mode: host  # port:8080 
    volumes:
      - ./config:/app/config:ro # add luna.yml to config folder
      # If you have custom assets (CSS/JS) that you want to test without rebuilding
      # - ./assets:/app/assets:ro
      # Optionally, also mount docker socket if you want to use the docker containers widget
      # - /var/run/docker.sock:/var/run/docker.sock:ro
    environment:
      - TZ=Etc/UTC
      - luna_CONFIG=/app/luna.yml
    env_file:
      - ./.env # create touch .env before deploy the container
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
<br>
</details>

<details>
<summary><strong>Docker container with Apprise Notification</strong></summary>


Create a new directory called `config` and add [`luna.yml`](docs/luna.yml) file in the directory


```bash
mkdir config && wget -O config/luna.yml https://raw.githubusercontent.com/luna-page/luna/refs/heads/main/docs/luna.yml
```


```yaml
services:
  apprise:
    image: linuxserver/apprise:latest
    container_name: apprise-api
    restart: unless-stopped
    ports:
      - "8000:8000"
    environment:
      - TZ=Etc/UTC

  luna:
    image: ghcr.io/luna-page/luna:main
    container_name: luna
    restart: unless-stopped
    # If you need luna to see services running directly on the host (e.g. DNS on port 53) and to proper use the notification system
    network_mode: host  # port:8080 
    volumes:
      - ./config:/app/config:ro # add luna.yml to config folder
      # If you have custom assets (CSS/JS) that you want to test without rebuilding
      # - ./assets:/app/assets:ro
      # Optionally, also mount docker socket if you want to use the docker containers widget
      # - /var/run/docker.sock:/var/run/docker.sock:ro
    environment:
      - TZ=Etc/UTC
      - luna_CONFIG=/app/luna.yml
    env_file:
      - ./.env # create touch .env before deploy the container
    # Important for SSE (Live Events) data flow
    logging:
      driver: "json-file"
      options:
        max-size: "10mb"
        max-file: "3"
    depends_on:
      - apprise
```

```bash
docker compose up -d
```
<br>
</details>


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
** To update the lates build just type `update`.

## Building from source

Choose one of the following methods:

<details>
<summary><strong>Build binary with Go</strong></summary>
<br>

Requirements: [Go](https://go.dev/dl/) >= v1.23

To build the project for your current OS and architecture, run:

```bash
go build -o build/luna .
```

To build for a specific OS and architecture, run:

```bash
GOOS=linux GOARCH=amd64 go build -o build/luna .
```

[*click here for a full list of GOOS and GOARCH combinations*](https://go.dev/doc/install/source#:~:text=$GOOS%20and%20$GOARCH)

Alternatively, if you just want to run the app without creating a binary, like when you're testing out changes, you can run:

```bash
go run .
```
<hr>
</details>

<details>
<summary><strong>Build project and Docker image with Docker</strong></summary>
<br>

Requirements: [Docker](https://docs.docker.com/engine/install/)

To build the project and image using just Docker, run:

*(replace `owner` with your name or organization)*

```bash
docker build -t owner/luna:latest .
```

If you wish to push the image to a registry (by default Docker Hub), run:

```bash
docker push owner/luna:latest
```

<hr>
</details>


