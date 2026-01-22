#!/bin/bash
set -euo pipefail

VERSION="${VERSION:-e2e-tests}"
BUILD_DATE="${BUILD_DATE:-$(date --utc +%Y-%m-%dT%H:%M:%SZ)}"

IMAGE="localhost:5001/adguardhome-sync:e2e"

docker build -f Dockerfile \
  --build-arg VERSION="${VERSION}" \
  --build-arg BUILD="${BUILD_DATE}" \
  -t "${IMAGE}" \
  .

docker push "${IMAGE}"
