#!/bin/bash

kubectl wait --for=jsonpath='{.status.phase}'=Succeeded pod/adguardhome-sync --timeout=1m
RESULT=$?
if [[ "${RESULT}" != "0" ]]; then
  kubectl logs adguardhome-sync
fi
exit ${RESULT}
