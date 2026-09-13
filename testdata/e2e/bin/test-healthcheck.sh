#!/bin/bash
set -e

MODE="${1:-env}"
PROTOCOL="${2:-http}"

echo "## Healthcheck CLI (${MODE} / ${PROTOCOL})" >> $GITHUB_STEP_SUMMARY
echo '```' >> $GITHUB_STEP_SUMMARY

if [[ "${MODE}" == "file" ]]; then
  CONFIG_FLAG="--config /etc/go/adguardhome-sync/config.yaml"
else
  CONFIG_FLAG=""
fi

echo ">> Check liveness via healthcheck CLI"
kubectl exec adguardhome-sync -- /opt/go/adguardhome-sync healthcheck ${CONFIG_FLAG} >> $GITHUB_STEP_SUMMARY

echo ">> Check liveness via health alias"
kubectl exec adguardhome-sync -- /opt/go/adguardhome-sync health ${CONFIG_FLAG} >> $GITHUB_STEP_SUMMARY

echo ">> Check readiness via healthcheck --ready"
kubectl exec adguardhome-sync -- /opt/go/adguardhome-sync healthcheck --ready ${CONFIG_FLAG} >> $GITHUB_STEP_SUMMARY

echo ">> Check healthcheck with custom --port"
if [[ "${PROTOCOL}" == "https" ]]; then
  kubectl exec adguardhome-sync -- /opt/go/adguardhome-sync healthcheck --port 9090 -k >> $GITHUB_STEP_SUMMARY
else
  kubectl exec adguardhome-sync -- /opt/go/adguardhome-sync healthcheck --port 9090 >> $GITHUB_STEP_SUMMARY
fi

echo ">> Check healthcheck with custom --url"
kubectl exec adguardhome-sync -- /opt/go/adguardhome-sync healthcheck --url "${PROTOCOL}://localhost:9090/livez" -k >> $GITHUB_STEP_SUMMARY
kubectl exec adguardhome-sync -- /opt/go/adguardhome-sync healthcheck --url "${PROTOCOL}://localhost:9090/readyz" -k >> $GITHUB_STEP_SUMMARY

echo '```' >> $GITHUB_STEP_SUMMARY
