#!/bin/bash
set -e

for pod in $(kubectl get pods -l bakito.net/adguardhome-sync=replica -o name); do
  echo "## Pod ${pod} logs" >> $GITHUB_STEP_SUMMARY
  echo '```' >> $GITHUB_STEP_SUMMARY
  kubectl logs ${pod} >> $GITHUB_STEP_SUMMARY
  ERRORS=$(kubectl logs ${pod} | grep '\[error\]' | wc -l)
  echo '```' >> $GITHUB_STEP_SUMMARY
  echo "Found ${ERRORS} error(s) in ${pod} log" >> $GITHUB_STEP_SUMMARY
  echo "----------------------------------------------" >> $GITHUB_STEP_SUMMARY

done
