#!/usr/bin/env bash
set -euo pipefail

if [ $# -eq 0 ]; then
    echo "Usage: run-tests-container.sh [tag]"
    exit 1
fi

TAG="$1"
NETWORK="pokemon-cache-test-net-${TAG}"

# The Docker socket is mounted (not Docker-in-Docker) so testcontainers-go,
# running inside this container, talks to the *host's* daemon and spins up
# sibling containers (Postgres, Redis). This container and every sibling
# container join a dedicated per-run bridge network (created below) and
# address each other by container alias on that network's internal ports,
# rather than this container using --network=host to reach the siblings'
# host-published ports. See test/integration/network_helper_test.go for the
# corresponding container-side wiring.
#
# The repo root is bind-mounted over the container's /src so report files
# (coverage.out, unit-test-report.json, integration-test-report.json) land
# back on the host, matching what platform-standard's Test stage does in CI.
docker network create "$NETWORK" >/dev/null
trap 'docker network rm "$NETWORK" >/dev/null 2>&1 || true' EXIT

docker run --rm \
    --network="$NETWORK" \
    -e TEST_NETWORK="$NETWORK" \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v "$(pwd)":/src \
    pokemon-cache-service-test-runner:"$TAG"
