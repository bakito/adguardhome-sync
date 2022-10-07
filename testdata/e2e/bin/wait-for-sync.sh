#!/bin/bash
set -e

kubectl wait --for=jsonpath='{.status.phase}'=Succeeded pod/adguardhome-sync --timeout=1m
