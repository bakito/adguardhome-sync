#!/bin/bash
set -e

echo "## Replica Features (${1:-mode})" >> $GITHUB_STEP_SUMMARY
echo '```' >> $GITHUB_STEP_SUMMARY

echo ">> Verify disabled features logged for replica 2 (v0.107.79)"
kubectl logs pod/adguardhome-sync | grep "Disabled features" >> $GITHUB_STEP_SUMMARY

echo '```' >> $GITHUB_STEP_SUMMARY
