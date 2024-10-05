#!/bin/bash
set -e

for pod in $(kubectl get pods -l bakito.net/adguardhome-sync=replica -o name); do
  echo "## Pod ${pod} logs" >> $GITHUB_STEP_SUMMARY
  echo '```' >> $GITHUB_STEP_SUMMARY
  K8S_LOGS=$(kubectl logs ${pod})
  # ignore certain errors
  LOGS=$(echo -e "${K8S_LOGS}" |
    grep -v -e "error.* deleting filter .* no such file or directory" |
    grep -v -e '\[error\] storage: recovered from panic: runtime' # https://github.com/AdguardTeam/AdGuardHome/issues/7315
  )

  echo -e "${K8S_LOGS}" >> $GITHUB_STEP_SUMMARY
  ERRORS=$(echo -e "${LOGS}"} | grep '\[error\]' | wc -l)
  TOTAL_ERRORS=$(echo -e "${K8S_LOGS}"} | grep '\[error\]' | wc -l)
  IGNORED_ERRORS=$(echo "${TOTAL_ERRORS} - ${ERRORS}" | bc)
  echo '```' >> $GITHUB_STEP_SUMMARY
  echo "Found ${ERRORS} error(s) (${IGNORED_ERRORS} ignored) in ${pod} log" >> $GITHUB_STEP_SUMMARY
  echo "----------------------------------------------" >> $GITHUB_STEP_SUMMARY
done
