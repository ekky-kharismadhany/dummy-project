#!/usr/bin/env bash
set -uo pipefail

if [ $# -eq 0 ]; then
    echo "Usage: cleanup-test-image.sh [tag]"
    exit 0
fi

TAG="$1"

docker rmi pokemon-cache-service-test-runner:"$TAG" || true
