//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"

	"pokemon-cache-service/internal/cache"
	"pokemon-cache-service/internal/pokeapi"
)

func TestRedisCache_SetAndGet(t *testing.T) {
	ctx := context.Background()

	testNet := testNetwork()
	const redisAlias = "pokemon-cache-it-redis"

	var redisOpts []testcontainers.ContainerCustomizer
	if testNet != "" {
		redisOpts = append(redisOpts,
			tcnetwork.WithNetworkName([]string{redisAlias}, testNet),
			// Same rationale as postgres in repository_test.go: the
			// module's default wait strategy dials the mapped host port,
			// unreachable from this test binary's own network namespace
			// when it's running inside a container on testNet.
			testcontainers.WithWaitStrategy(
				wait.ForListeningPort("6379/tcp").WithStartupTimeout(10*time.Second).SkipExternalCheck(),
				wait.ForLog("Ready to accept connections"),
			),
		)
	}

	redisContainer, err := redis.Run(ctx, "redis:7-alpine", redisOpts...)
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}
	t.Cleanup(func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate redis container: %v", err)
		}
	})

	var addr string
	if testNet != "" {
		addr = fmt.Sprintf("%s:6379", redisAlias)
	} else {
		addr, err = redisContainer.Endpoint(ctx, "")
		if err != nil {
			t.Fatalf("failed to get redis endpoint: %v", err)
		}
	}

	client := goredis.NewClient(&goredis.Options{Addr: addr})
	defer client.Close()

	pokemonCache := cache.NewRedisCache(client)

	data := &pokeapi.PokemonData{Name: "snorlax", Height: 21, Weight: 4600, BaseExperience: 189}

	if err := pokemonCache.Set(ctx, "snorlax", data, time.Minute); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, hit, err := pokemonCache.Get(ctx, "snorlax")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !hit {
		t.Fatal("expected cache hit")
	}
	if *got != *data {
		t.Errorf("expected %+v, got %+v", data, got)
	}

	_, miss, err := pokemonCache.Get(ctx, "unknown-pokemon")
	if err != nil {
		t.Fatalf("Get for missing key failed: %v", err)
	}
	if miss {
		t.Error("expected cache miss for unknown key")
	}
}
