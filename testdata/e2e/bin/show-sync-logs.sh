#!/bin/bash
set -e

echo "## Pod adguardhome-sync logs" >> $GITHUB_STEP_SUMMARY
echo '```' >> $GITHUB_STEP_SUMMARY
kubectl logs adguardhome-sync  >> $GITHUB_STEP_SUMMARY
echo '```' >> $GITHUB_STEP_SUMMARY
ERRORS=$(kubectl logs adguardhome-sync | grep Error | wc -l)
echo "Found ${ERRORS} error(s) in adguardhome-sync log";  >> $GITHUB_STEP_SUMMARY
if [[ "${ERRORS}" != "0" ]]; then exit 1; fi
