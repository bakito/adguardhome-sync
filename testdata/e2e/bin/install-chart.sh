#!/bin/bash
set -e

MODE="${1:-env}"
SYNC_ENABLED="${SYNC_ENABLED:-true}"

kubectl config set-context --current --namespace=agh-e2e 2>/dev/null || true

if [[ $(helm list --no-headers -n agh-e2e 2>/dev/null | grep -w agh-e2e | wc -l) == "1" ]]; then
  helm delete agh-e2e -n agh-e2e --wait
fi
helm install agh-e2e testdata/e2e -n agh-e2e --create-namespace --set mode="${MODE}" --set sync.enabled="${SYNC_ENABLED}"

