#!/bin/bash
set -e

echo "## Pod adguardhome-origin ${1} logs" >> $GITHUB_STEP_SUMMARY
echo '```' >> $GITHUB_STEP_SUMMARY
kubectl logs adguardhome-origin >> $GITHUB_STEP_SUMMARY
echo '```' >> $GITHUB_STEP_SUMMARY
