#!/bin/bash
set -e

echo "wait another scrape interval (30s)"

sleep 30

echo "## Pod adguardhome-sync metrics" >> $GITHUB_STEP_SUMMARY
echo '```' >> $GITHUB_STEP_SUMMARY
curl http://localhost:9090/metrics -s >> $GITHUB_STEP_SUMMARY
echo '```' >> $GITHUB_STEP_SUMMARY
