[![Go](https://github.com/bakito/adguardhome-sync/actions/workflows/go.yml/badge.svg)](https://github.com/bakito/adguardhome-sync/actions/workflows/go.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/bakito/adguardhome-sync)](https://goreportcard.com/report/github.com/bakito/adguardhome-sync)

# AdGuardHome sync

Synchronize [AdGuardHome](https://github.com/AdguardTeam/AdGuardHome) config to a replica instance.

This is still work in progress. Changes on how parameters are provided will change.

## Current sync features

- Filters
- Rewrites
- Services
- Clients

## Install

```bash
go get -u github.com/bakito/adguardhome-sync
```

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