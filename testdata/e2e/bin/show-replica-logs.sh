#!/bin/bash
set -e

for pod in $(kubectl get pods -l bakito.net/adguardhome-sync=replica -o name); do
  echo Pod "${pod} logs"
  kubectl logs ${pod}
  ERRORS=$(kubectl logs ${pod} | grep '\[error\]' | wc -l)
  echo "Found ${ERRORS} error(s) in log"
  echo "----------------------------------------------"
done
