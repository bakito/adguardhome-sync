#!/bin/bash
set -e

kubectl config set-context --current --namespace=agh-e2e

if [[ $(helm list --no-headers -n agh-e2e | grep agh-e2e | wc -l) == "1" ]]; then
  helm delete agh-e2e -n agh-e2e --wait
fi
helm install agh-e2e testdata/e2e -n agh-e2e --create-namespace
