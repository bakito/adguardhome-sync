[![Go](https://github.com/bakito/adguardhome-sync/actions/workflows/go.yml/badge.svg)](https://github.com/bakito/adguardhome-sync/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bakito/adguardhome-sync)](https://goreportcard.com/report/github.com/bakito/adguardhome-sync)
[![Coverage Status](https://coveralls.io/repos/github/bakito/adguardhome-sync/badge.svg?branch=main)](https://coveralls.io/github/bakito/adguardhome-sync?branch=main)

# AdGuardHome sync

Synchronize [AdGuardHome](https://github.com/AdguardTeam/AdGuardHome) config to a replica instance.

## Current sync features

- General Settings
- Filters
- Rewrites
- Services
- Clients
- DNS Config
- DHCP Config

By default, all features are enabled. Single features can be disabled in the config.

### Setup of initial instances

New AdGuardHome replica instances can be automatically installed if enabled via the config autoSetup. During automatic
installation, the admin interface will be listening on port 3000 in runtime.

To skip automatic setup

## Install

Get from [releases](https://github.com/bakito/adguardhome-sync/releases) or install from source

```bash
go install github.com/bakito/adguardhome-sync@latest
```

## Prerequisites

Both the origin instance must be initially setup via the AdguardHome installation wizard.

## Run

```bash

export ORIGIN_URL=https://192.168.1.2:3000
export ORIGIN_USERNAME=username
export ORIGIN_PASSWORD=password
export REPLICA_URL=http://192.168.1.3
export REPLICA_USERNAME=username
export REPLICA_PASSWORD=password

# run once
adguardhome-sync run

# run as daemon
adguardhome-sync run --cron "*/10 * * * *"
```

## docker cli

```bash
docker run -d \
  --name=adguardhome-sync \
  -p 8080:8080 \
  -v /path/to/appdata/config/adguardhome-sync.yaml:/config/adguardhome-sync.yaml \
  --restart unless-stopped \
  ghcr.io/bakito/adguardhome-sync:latest
```

## docker compose

### config file

```yaml
---
version: "2.1"
services:
  adguardhome-sync:
    image: ghcr.io/bakito/adguardhome-sync
    container_name: adguardhome-sync
    command: run --config /config/adguardhome-sync.yaml
    volumes:
      - /path/to/appdata/config/adguardhome-sync.yaml:/config/adguardhome-sync.yaml
    ports:
      - 8080:8080
    restart: unless-stopped
```

### env

```yaml
---
version: "2.1"
services:
  adguardhome-sync:
    image: ghcr.io/bakito/adguardhome-sync
    container_name: adguardhome-sync
    command: run
    environment:
      - ORIGIN_URL=https://192.168.1.2:3000
      - ORIGIN_USERNAME=username
      - ORIGIN_PASSWORD=password
      - REPLICA_URL=http://192.168.1.3
      - REPLICA_USERNAME=username
      - REPLICA_PASSWORD=password
      - REPLICA1_URL=http://192.168.1.4
      - REPLICA1_USERNAME=username
      - REPLICA1_PASSWORD=password
      - REPLICA1_APIPATH=/some/path/control
      # - REPLICA1_AUTOSETUP=true # if true, AdGuardHome is automatically initialized. 
      - CRON=*/10 * * * * # run every 10 minutes
      - RUNONSTART=true
      # Configure sync features; by default all features are enabled.
      # - FEATURES_GENERALSETTINGS=true
      # - FEATURES_QUERYLOGCONFIG=true
      # - FEATURES_STATSCONFIG=true
      # - FEATURES_CLIENTSETTINGS=true
      # - FEATURES_SERVICES=true
      # - FEATURES_FILTERS=true
      # - FEATURES_DHCP_SERVERCONFIG=true
      # - FEATURES_DHCP_STATICLEASES=true
      # - FEATURES_DNS_SERVERCONFIG=true
      # - FEATURES_DNS_ACCESSLISTS=true
      # - FEATURES_DNS_REWRITES=true
    ports:
      - 8080:8080
    restart: unless-stopped
```

### Config file

location: $HOME/.adguardhome-sync.yaml

```yaml
# cron expression to run in daemon mode. (default; "" = runs only once)
cron: "*/10 * * * *"

# runs the synchronisation on startup
runOnStart: true

origin:
  # url of the origin instance
  url: https://192.168.1.2:3000
  # apiPath: define an api path if other than "/control"
  # insecureSkipVerify: true # disable tls check
  username: username
  password: password

# replica instance (optional, if only one)
replica:
  # url of the replica instance
  url: http://192.168.1.3
  username: username
  password: password

# replicas instances (optional, if more than one)
replicas:
  # url of the replica instance
  - url: http://192.168.1.3
    username: username
    password: password
  - url: http://192.168.1.4
    username: username
    password: password
    # autoSetup: true # if true, AdGuardHome is automatically initialized. 

# Configure the sync API server, disabled if api port is 0
api:
  # Port, default 8080
  port: 8080
  # if username and password are defined, basic auth is applied to the sync API 
  username: username
  password: password

# Configure sync features; by default all features are enabled.
features:
  generalSettings: true
  queryLogConfig: true
  statsConfig: true
  clientSettings: true
  services: true
  filters: true
  dhcp:
    serverConfig: true
    staticLeases: true
  dns:
    serverConfig: true
    accessLists: true
    rewrites: true
```

## Log Level

The log level can be set with the environment variable: LOG_LEVEL

The following log levels are supported (default: info)

- debug
- info
- warn
- error
