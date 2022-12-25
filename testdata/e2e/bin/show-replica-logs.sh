#!/bin/bash
set -e

for pod in $(kubectl get pods -l bakito.net/adguardhome-sync=replica -o name); do
  echo "## Pod ${pod} logs" >> $GITHUB_STEP_SUMMARY
  echo '```' >> $GITHUB_STEP_SUMMARY
  LOGS=$(kubectl logs ${pod})
  # ignore certain errors
  LOGS=$(echo -e "${LOGS}" | grep -v -e "error.* deleting filter .* no such file or directory" )
  # https://github.com/AdguardTeam/AdGuardHome/issues/4944
  LOGS=$(echo -e "${LOGS}" | grep -v -e "error.* creating dhcpv4 srv")
  echo -e "${LOGS}" >> $GITHUB_STEP_SUMMARY
  ERRORS=$(echo -e "${LOGS}"} | grep '\[error\]' | wc -l)
  echo '```' >> $GITHUB_STEP_SUMMARY
  echo "Found ${ERRORS} error(s) in ${pod} log" >> $GITHUB_STEP_SUMMARY
  echo "----------------------------------------------" >> $GITHUB_STEP_SUMMARY

done
