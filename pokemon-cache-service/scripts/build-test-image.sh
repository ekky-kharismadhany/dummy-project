#!/usr/bin/env bash
set -euo pipefail

if [ $# -eq 0 ]; then
    echo "Usage: build-test-image.sh [tag]"
    exit 1
fi

TAG="$1"

DOCKER_BUILDKIT=1 docker build \
    -t pokemon-cache-service-test-runner:"$TAG" \
    -f Dockerfile.test .
