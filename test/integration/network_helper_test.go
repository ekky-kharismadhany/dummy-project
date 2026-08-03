//go:build integration

package integration

import "os"

// testNetworkEnvVar names a pre-existing Docker network that sibling
// containers (postgres, redis) should join instead of relying on published
// host ports. It's set by scripts/run-tests.sh when this test binary itself
// runs inside a container attached to that same network via a mounted
// Docker socket (see Dockerfile.test) — a container on a plain bridge
// network has no route to another container's host-mapped port, so
// testcontainers-go's default "dial the mapped host port" wait strategy and
// connection string can't be used in that mode. It's empty for a plain
// local `go test -tags=integration`, where Host()+MappedPort() is correct.
const testNetworkEnvVar = "TEST_NETWORK"

func testNetwork() string {
	return os.Getenv(testNetworkEnvVar)
}
