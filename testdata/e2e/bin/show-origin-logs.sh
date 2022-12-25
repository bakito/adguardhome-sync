#!/bin/bash
set -e

echo "## Pod adguardhome-origin logs" >> $GITHUB_STEP_SUMMARY
echo '```' >> $GITHUB_STEP_SUMMARY
kubectl logs adguardhome-origin >> $GITHUB_STEP_SUMMARY
echo '```' >> $GITHUB_STEP_SUMMARY
