.PHONY: test test-integration test-docker mock build

TEST_TAG ?= local

test:
	go test ./... -short

test-integration:
	go test ./test/integration/... -tags=integration

# Runs unit + integration tests in Docker (docker-cli inside the container
# talks to the host daemon via the mounted socket to spin up sibling
# Postgres/Redis containers on a dedicated per-run network). See
# scripts/run-tests.sh and Dockerfile.test.
test-docker:
	./scripts/build-test-image.sh $(TEST_TAG)
	./scripts/run-tests-container.sh $(TEST_TAG)
	./scripts/cleanup-test-image.sh $(TEST_TAG)

# Regenerates gomock mocks for the repository/cache/pokeapi interfaces.
mock:
	go generate ./...

build:
	go build -o bin/server ./cmd/server
