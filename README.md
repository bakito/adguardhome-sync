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

Both the origin instance and replica(s) must be initially set up with AdguardHome via the AdguardHome installation wizard.

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

set FEATURES_DHCP_SERVER_CONFIG=false
set FEATURES_DHCP_STATIC_LEASES=false

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

## Config via environment variables

For Replicas replace `#` with the index number for the replica. E.g: `REPLICA#_URL` -> `REPLICA1_URL`

| Name                                 | Type   | Description                                               |
|:-------------------------------------|--------|:----------------------------------------------------------|
| ORIGIN_URL (string)                  | string | URL of adguardhome instance                               |
| ORIGIN_WEB_URL (string)              | string | Web URL of adguardhome instance                           |
| ORIGIN_API_PATH (string)             | string | API Path                                                  |
| ORIGIN_USERNAME (string)             | string | Adguardhome username                                      |
| ORIGIN_PASSWORD (string)             | string | Adguardhome password                                      |
| ORIGIN_COOKIE (string)               | string | Adguardhome cookie                                        |
| ORIGIN_REQUEST_HEADERS (map)         | map    | Request Headers 'key1:value1,key2:value2'                 |
| ORIGIN_INSECURE_SKIP_VERIFY (bool)   | bool   | Skip TLS verification                                     |
| ORIGIN_AUTO_SETUP (bool)             | bool   | Automatically setup the instance if it is not initialized |
| ORIGIN_INTERFACE_NAME (string)       | string | Network interface name                                    |
| ORIGIN_DHCP_SERVER_ENABLED (bool)    | bool   | Enable DHCP server                                        |
| REPLICA#_URL (string)                | string | URL of adguardhome instance                               |
| REPLICA#_WEB_URL (string)            | string | Web URL of adguardhome instance                           |
| REPLICA#_API_PATH (string)           | string | API Path                                                  |
| REPLICA#_USERNAME (string)           | string | Adguardhome username                                      |
| REPLICA#_PASSWORD (string)           | string | Adguardhome password                                      |
| REPLICA#_COOKIE (string)             | string | Adguardhome cookie                                        |
| REPLICA#_REQUEST_HEADERS (map)       | map    | Request Headers 'key1:value1,key2:value2'                 |
| REPLICA#_INSECURE_SKIP_VERIFY (bool) | bool   | Skip TLS verification                                     |
| REPLICA#_AUTO_SETUP (bool)           | bool   | Automatically setup the instance if it is not initialized |
| REPLICA#_INTERFACE_NAME (string)     | string | Network interface name                                    |
| REPLICA#_DHCP_SERVER_ENABLED (bool)  | bool   | Enable DHCP server                                        |
| CRON (string)                        | string | Cron expression for the sync interval                     |
| RUN_ON_START (bool)                  | bool   | Run the sung on startup                                   |
| PRINT_CONFIG_ONLY (bool)             | bool   | Print current config only and stop the application        |
| CONTINUE_ON_ERROR (bool)             | bool   | Continue sync on errors                                   |
| API_PORT (int)                       | int    | API port                                                  |
| API_USERNAME (string)                | string | API username                                              |
| API_PASSWORD (string)                | string | API password                                              |
| API_DARK_MODE (bool)                 | bool   | API dark mode                                             |
| API_METRICS_ENABLED (bool)           | bool   | Enable metrics                                            |
| API_METRICS_SCRAPE_INTERVAL (int64)  | int64  | Interval for metrics scraping                             |
| API_METRICS_QUERY_LOG_LIMIT (int)    | int    | Metrics log query limit                                   |
| API_TLS_CERT_DIR (string)            | string | API TLS certificate directory                             |
| API_TLS_CERT_NAME (string)           | string | API TLS certificate file name                             |
| API_TLS_KEY_NAME (string)            | string | API TLS key file name                                     |
| FEATURES_DNS_ACCESS_LISTS (bool)     | bool   | Sync DNS access lists                                     |
| FEATURES_DNS_SERVER_CONFIG (bool)    | bool   | Sync DNS server config                                    |
| FEATURES_DNS_REWRITES (bool)         | bool   | Sync DNS rewrites                                         |
| FEATURES_DHCP_SERVER_CONFIG (bool)   | bool   | Sync DHCP server config                                   |
| FEATURES_DHCP_STATIC_LEASES (bool)   | bool   | Sync DHCP static leases                                   |
| FEATURES_GENERAL_SETTINGS (bool)     | bool   | Sync general settings                                     |
| FEATURES_QUERY_LOG_CONFIG (bool)     | bool   | Sync query log config                                     |
| FEATURES_STATS_CONFIG (bool)         | bool   | Sync stats config                                         |
| FEATURES_CLIENT_SETTINGS (bool)      | bool   | Sync client settings                                      |
| FEATURES_SERVICES (bool)             | bool   | Sync services                                             |
| FEATURES_FILTERS (bool)              | bool   | Sync filters                                              |
| FEATURES_THEME (bool)                | bool   | Sync the weg UI theme                                     |

### Unraid

⚠️ Disclaimer: Tere exists an unraid tepmlate for this application. This template is not managed by this project.
Also, as unraid is not known to me, I can not give any support on unraind templates.

Note when running the Docker container in Unraid please remove unneeded env variables if don't needed.
If replica2 isn't used this can cause sync errors.

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
  # requestHeaders: # Additional request headers
  #   AAA: bbb

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
    # requestHeaders: # Additional request headers
    #   AAA: bbb

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
  #   enabled: true
  #   scrapeInterval: 30s 
  #   queryLogLimit: 10000

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

## Home Assistant AdGuard Home Add-on users

To enable syncing with a Home Assistant instance using
the [AdGuard Home Add-on](https://github.com/hassio-addons/addon-adguard-home), you will need to enable the disabled
ports, under the Network heading

![show-disabled-ports](https://github.com/user-attachments/assets/1df5f352-37a2-4508-82ec-7f270087d0b4)

And then set the port of your choice for the Web interface

![web-interface-port](https://github.com/user-attachments/assets/286ed030-4831-4f49-8b29-53e8802129c3)

Don't forget to save and restart the add-on.

Depending on your setup, you may also need to disable SSL for the add-on.

The username:password required for the Home Assistant replica is the one you use to login to your instance, however it's
recommended to setup a new local only user with minimal permissions.

All credit for this method goes to [Brunty](https://github.com/brunty) who has a far
more [detailed write up](https://brunty.me/post/replicate-adguard-home-settings-into-home-assistant-adguard-home-addon/)
about this on his blog.

## Log Level

The log level can be set with the environment variable: `LOG_LEVEL`

The following log levels are supported (default: info)

- debug
- info
- warn
- error

## Log Format

Default log format is `console`.
It can be changed to `json` by setting the environment variable: `LOG_FORMAT=json`.

## Video Tutorials

- [Como replicar la configuración de tu servidor DNS Adguard automáticamente - Tu servidor Part #12](https://www.youtube.com/watch?v=1LPeu_JG064) (
  Spanish) by [Jonatan Castro](https://github.com/jcastro)
