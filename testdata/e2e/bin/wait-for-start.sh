#!/bin/bash

kubectl wait --for=jsonpath='{.status.phase}'=Running pod/adguardhome-sync --timeout=1m
kubectl describe pod/adguardhome-sync
kubectl logs pod/adguardhome-sync
