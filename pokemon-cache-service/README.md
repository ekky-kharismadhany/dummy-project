# pokemon-cache-service

A small Go HTTP service demonstrating three distinct dependency types —
PostgreSQL, Redis, and a third-party HTTP API (PokéAPI) — used as a test
consumer of the `platform-standard` Jenkins shared library.

`GET /pokemon/{name}` checks Redis first; on a miss it calls
[PokéAPI](https://pokeapi.co/docs/v2), caches the result for 5 minutes, and
logs the lookup to Postgres. Cache and log failures are logged but never fail
the request.

```json
{
  "name": "pikachu",
  "height": 4,
  "weight": 60,
  "base_experience": 112,
  "source": "cache"
}
```

## Running locally

```bash
docker-compose up -d postgres redis

export POSTGRES_DSN="postgres://postgres:postgres@localhost:5432/pokemon?sslmode=disable"
export REDIS_ADDR="localhost:6379"
export POKEAPI_BASE_URL="https://pokeapi.co/api/v2"

go run ./cmd/server
# curl http://localhost:8080/pokemon/pikachu
```

All three env vars default to the localhost values above, so `go run
./cmd/server` works out of the box once `docker-compose up` has started
Postgres and Redis.

## Running unit tests

```bash
make test
```

Runs `go test ./... -short`. All three dependencies (repository, cache,
PokéAPI client) are mocked with [go.uber.org/mock](https://github.com/uber-go/mock)
(gomock) — no network access or Docker required. Mocks live under
`internal/mock/{cache,repository,pokeapi}` and are generated from the
`//go:generate mockgen ...` directive next to each interface; regenerate them
with `make mock` after changing an interface.

## Running integration tests

```bash
make test-integration
```

Runs `go test ./test/integration/... -tags=integration`. Requires Docker:
spins up real `postgres:16-alpine` and `redis:7-alpine` containers via
testcontainers-go (Ryuk left enabled so containers are cleaned up even if
the run is aborted), plus a local `httptest.Server` stub standing in for
PokéAPI so tests never hit the real, rate-limited public API.

### Running everything in Docker

```bash
make test-docker
```

Builds `Dockerfile.test` and runs both the unit and integration suites
inside a single container, driven by `scripts/run-tests.sh`. `docker-cli` is
installed in that image and the container is run with the host's Docker
socket mounted (not Docker-in-Docker), so testcontainers-go — running
*inside* the container — talks to the host daemon and spins up sibling
Postgres/Redis containers on a dedicated per-run bridge network
(`scripts/run-tests-container.sh`). The test binary and its siblings address
each other by container alias on that network rather than through
host-mapped ports; see `test/integration/network_helper_test.go`'s
`TEST_NETWORK` handling for the container-side wiring. This mirrors the
Dockerized test-runner pattern used in `map-service-go`.

## CI/CD

This repo's `Jenkinsfile` contains no pipeline logic of its own — it only
calls the shared library, pinned to an immutable tag:

```groovy
@Library('platform-standard@v1') _
standardPipeline(
    serviceName: 'pokemon-cache-service',
    nodeLabel: 'slave-01',
    agentWorkspacePattern: 'workspace/${BRANCH_NAME}/src/git.bluebird.id/platform/pokemon-cache-service'
)
```

Checkout, Unit Test, Integration Test, and Build Image are all defined in
`platform-standard`. See that library's README for stage details. Deploy and
SonarQube stages are intentionally not yet part of `standardPipeline()` and
will be added once ArgoCD/SonarQube integration is defined for this pipeline.
