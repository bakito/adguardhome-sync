[![Go](https://github.com/bakito/adguardhome-sync/actions/workflows/go.yml/badge.svg)](https://github.com/bakito/adguardhome-sync/actions/workflows/go.yml)
[![e2e tests](https://github.com/bakito/adguardhome-sync/actions/workflows/e2e.yaml/badge.svg)](https://github.com/bakito/adguardhome-sync/actions/workflows/e2e.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bakito/adguardhome-sync)](https://goreportcard.com/report/github.com/bakito/adguardhome-sync)
[![Coverage Status](https://coveralls.io/repos/github/bakito/adguardhome-sync/badge.svg?branch=main&service=github)](https://coveralls.io/github/bakito/adguardhome-sync?branch=main)

# AdGuardHome sync

Synchronize [AdGuardHome](https://github.com/AdguardTeam/AdGuardHome) config to replica instances.

## FAQ & Deprecations

Please check the wiki
for [FAQ](https://github.com/bakito/adguardhome-sync/wiki/FAQ)
and [Deprecations](https://github.com/bakito/adguardhome-sync/wiki/Deprecations).

## Current sync features

- General Settings
- Filters
- Rewrites
- Services
- Clients
- DNS Config
- DHCP Config
- Theme

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

## Username / Password vs. Cookie

Some instances of AdGuard Home do not support basic authentication. For instance, many routers with built-in Adguard
Home support do not. If this is the case, a valid cookie may be provided instead. If the router protects the AdGuard
instance behind its own authentication, the cookie from an authenticated request may allow the sync to succeed.

- This has been tested successfully against GL.Inet routers with AdGuard Home.
- Note: due to the short validity of cookies, this approach is likely only suitable for one-time syncs

## Run Linux/Mac

```bash

export LOG_LEVEL=info
export ORIGIN_URL=https://192.168.1.2:3000
export ORIGIN_USERNAME=username
export ORIGIN_PASSWORD=password
# export ORIGIN_COOKIE=Origin-Cookie-Name=CCCOOOKKKIIIEEE
export REPLICA1_URL=http://192.168.1.3
export REPLICA1_USERNAME=username
export REPLICA1_PASSWORD=password
# export REPLICA_COOKIE=Replica-Cookie-Name=CCCOOOKKKIIIEEE

# run once
adguardhome-sync run

# run as daemon
adguardhome-sync run --cron "0 */2 * * *"
```

### Run as Linux Service via Systemd

> Verified on Ubuntu Linux 24.04

Assume you have downloaded the the `adguardhome-sync` binary to `/opt/adguardhome-sync`.

Create systemd service file `/opt/adguardhome-sync/adguardhome-sync.service`:

```
[Unit]
Description = AdGuardHome Sync
After = network.target

[Service]
ExecStart = /opt/adguardhome-sync/adguardhome-sync --config /opt/adguardhome-sync/adguardhome-sync.yaml run

[Install]
WantedBy = multi-user.target

```

Create a configuration file `/opt/adguardhome-sync/adguardhome-sync.yaml`, please follow [Config file](#config-file-1)
section below for details.

Install and enable service:

```bash
sudo cp /opt/adguardhome-sync/adguardhome-sync.service /etc/systemd/system/

sudo systemctl enable adguardhome-sync.service

sudo systemctl start adguardhome-sync.service

```

Then you can check the status:

```bash
sudo systemctl status adguardhome-sync.service

```

If web UI has been enabled in configuration (default port is 8080), can also check the status via
`http://<server-IP>:8080`

## Run Windows

```bash
@ECHO OFF
@TITLE AdGuardHome-Sync

REM set LOG_LEVEL=debug
set LOG_LEVEL=info
REM set LOG_LEVEL=warn
REM set LOG_LEVEL=error

set ORIGIN_URL=http://192.168.1.2:3000
set ORIGIN_USERNAME=username
set ORIGIN_PASSWORD=password
# set ORIGIN_COOKIE=Origin-Cookie-Name=CCCOOOKKKIIIEEE

set REPLICA1_URL=http://192.168.2.2:3000
set REPLICA1_USERNAME=username
set REPLICA1_PASSWORD=password
# set REPLICA1_COOKIE=Replica-Cookie-Name=CCCOOOKKKIIIEEE

set FEATURES_DHCP=false
set FEATURES_DHCP_SERVERCONFIG=false
set FEATURES_DHCP_STATICLEASES=false

# run once
adguardhome-sync run

# run as daemon
adguardhome-sync run --cron "0 */2 * * *"
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
      LOG_LEVEL: "info"
      ORIGIN_URL: "https://192.168.1.2:3000"
      # ORIGIN_WEB_URL: "https://some-other.url" # used in the web interface (default: <origin-url>

      ORIGIN_USERNAME: "username"
      ORIGIN_PASSWORD: "password"
      REPLICA1_URL: "http://192.168.1.3"
      REPLICA1_USERNAME: "username"
      REPLICA1_PASSWORD: "password"
      REPLICA2_URL: "http://192.168.1.4"
      REPLICA2_USERNAME: "username"
      REPLICA2_PASSWORD: "password"
      REPLICA2_API_PATH: "/some/path/control"
      # REPLICA2_WEB_URL: "https://some-other.url" # used in the web interface (default: <replica-url>
      # REPLICA2_AUTO_SETUP: true # if true, AdGuardHome is automatically initialized.
      # REPLICA2_INTERFACE_NAME: 'ens18' # use custom dhcp interface name
      # REPLICA2_DHCP_SERVER_ENABLED: true/false (optional) enables/disables the dhcp server on the replica
      CRON: "0 */2 * * *" # run every 2 hours
      RUN_ON_START: "true"
      # CONTINUE_ON_ERROR: false # If enabled, the synchronisation task will not fail on single errors, but will log the errors and continue

      # Configure the sync API server, disabled if api port is 0
      API_PORT: 8080
      # API_DARK_MODE: "true"
      # API_USERNAME: admin
      # API_PASSWORD: secret
      # the directory of the provided tls certs
      # API_TLS_CERT_DIR: /path/to/certs
      # the name of the cert file (default: tls.crt)
      # API_TLS_CERT_NAME: foo.crt
      # the name of the key file (default: tls.key)
      # API_TLS_KEY_NAME: bar.key

      # Configure sync features; by default all features are enabled.
      # FEATURES_GENERAL_SETTINGS: "true"
      # FEATURES_QUERY_LOG_CONFIG: "true"
      # FEATURES_STATS_CONFIG: "true"
      # FEATURES_CLIENT_SETTINGS: "true"
      # FEATURES_SERVICES: "true"
      # FEATURES_FILTERS: "true"
      # FEATURES_DHCP_SERVER_CONFIG: "true"
      # FEATURES_DHCP_STATIC_LEASES: "true"
      # FEATURES_DNS_SERVER_CONFIG: "true"
      # FEATURES_DNS_ACCESS_LISTS: "true"
      # FEATURES_DNS_REWRITES: "true"
      # FEATURES_THEME: "true" # if false the UI theme is not synced
    ports:
      - 8080:8080
    restart: unless-stopped
```

### Config file

location: $HOME/.adguardhome-sync.yaml

```yaml
# cron expression to run in daemon mode. (default; "" = runs only once)
cron: "0 */2 * * *"

# runs the synchronisation on startup
runOnStart: true

# If enabled, the synchronisation task will not fail on single errors, but will log the errors and continue
continueOnError: false

origin:
  # url of the origin instance
  url: https://192.168.1.2:3000
  # apiPath: define an api path if other than "/control"
  # insecureSkipVerify: true # disable tls check
  username: username
  password: password
  # cookie: Origin-Cookie-Name=CCCOOOKKKIIIEEE

# replicas instances
replicas:
  # url of the replica instance
  - url: http://192.168.1.3
    username: username
    password: password
    # cookie: Replica1-Cookie-Name=CCCOOOKKKIIIEEE
  - url: http://192.168.1.4
    username: username
    password: password
    # cookie: Replica2-Cookie-Name=CCCOOOKKKIIIEEE
    # autoSetup: true # if true, AdGuardHome is automatically initialized.
    # webURL: "https://some-other.url" # used in the web interface (default: <replica-url>

# Configure the sync API server, disabled if api port is 0
api:
  # Port, default 8080
  port: 8080
  # if username and password are defined, basic auth is applied to the sync API
  username: username
  password: password
  # enable api dark mode
  darkMode: true

 # enable metrics on path '/metrics' (api port must be != 0)
# metrics:
  # enabled: true
  # scrapeInterval: 30s 
  # queryLogLimit: 10000

# enable tls for the api server
# tls:
#   # the directory of the provided tls certs
#   certDir: /path/to/certs
#   # the name of the cert file (default: tls.crt)
#   certName: foo.crt
#   # the name of the key file (default: tls.key)
#   keyName: bar.key

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

The log level can be set with the environment variable: `LOG_LEVEL`

The following log levels are supported (default: info)

- debug
- info
- warn
- error

## Log Format

Default log format is `console`.
It can be changed to `json`by setting the environment variable: `LOG_FORMAT=json`

## Video Tutorials

- [Como replicar la configuración de tu servidor DNS Adguard automáticamente - Tu servidor Part #12](https://www.youtube.com/watch?v=1LPeu_JG064) (
  Spanish) by [Jonatan Castro](https://github.com/jcastro)
