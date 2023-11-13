#!/bin/bash
set -e
  echo "## AdGuardHome.yaml of latest replica" >> $GITHUB_STEP_SUMMARY
  echo '```' >> $GITHUB_STEP_SUMMARY
  kubectl exec adguardhome-replica-latest -- cat /opt/adguardhome/conf/AdGuardHome.yaml >> $GITHUB_STEP_SUMMARY
  echo '```' >> $GITHUB_STEP_SUMMARY
