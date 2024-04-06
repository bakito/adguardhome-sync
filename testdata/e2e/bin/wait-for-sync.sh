#!/bin/bash

kubectl wait --for=jsonpath='{.status.phase}'=Running pod/adguardhome-sync --timeout=1m

kubectl port-forward pod/adguardhome-sync 9090:9090 &

for i in {1..6}; do
  sleep 10
   RUNNING=$(curl ${1}://localhost:9090/api/v1/status -s -k | jq -r .syncRunning)
   echo "SyncRunning = ${RUNNING}"
   if [[ "${RUNNING}" == "false" ]]; then
     exit 0
   fi
done
