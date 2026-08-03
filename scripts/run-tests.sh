#!/usr/bin/env sh
# Runs inside Dockerfile.test at container *run* time (not build time), so
# the mounted Docker socket is available for testcontainers-go to launch
# sibling Postgres/Redis containers.
set -u

cd /src || exit 1

echo "Running unit tests (mocked)..."
go test ./... -short
UNIT_STATUS=$?

echo "Running integration tests (testcontainers-go)..."
go test ./test/integration/... -tags=integration -v
INTEGRATION_STATUS=$?

if [ "$UNIT_STATUS" -ne 0 ] || [ "$INTEGRATION_STATUS" -ne 0 ]; then
    echo "unit exit=$UNIT_STATUS integration exit=$INTEGRATION_STATUS"
    exit 1
fi
