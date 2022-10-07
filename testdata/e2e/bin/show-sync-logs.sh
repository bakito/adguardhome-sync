#!/bin/bash
set -e

echo Pod adguardhome-sync logs
kubectl logs adguardhome-sync
ERRORS=$(kubectl logs adguardhome-sync | grep Error | wc -l)
echo "Found ${ERRORS} error(s) in log";
if [[ "${ERRORS}" != "0" ]]; then exit 1; fi
