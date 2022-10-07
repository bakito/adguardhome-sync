#!/bin/bash
set -e

echo "wait for adguardhome pods"
for pod in $(kubectl get pods -l bakito.net/adguardhome-sync -o name); do
  kubectl wait --for condition=Ready ${pod} --timeout=30s
done
