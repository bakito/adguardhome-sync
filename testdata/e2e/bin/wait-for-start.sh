#!/bin/bash

kubectl wait --for=jsonpath='{.status.phase}'=Running pod/adguardhome-sync --timeout=1m
