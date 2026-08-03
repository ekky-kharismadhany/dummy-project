//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"

	"pokemon-cache-service/internal/repository"
)

func TestPostgresLookupLogRepository_LogLookup(t *testing.T) {
	ctx := context.Background()

	schema, err := os.ReadFile("../testdata/schema.sql")
	if err != nil {
		t.Fatalf("failed to read schema: %v", err)
	}

	testNet := testNetwork()
	const pgAlias = "pokemon-cache-it-pg"

	pgOpts := []testcontainers.ContainerCustomizer{
		postgres.WithDatabase("pokemon"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	}
	if testNet != "" {
		pgOpts = append(pgOpts,
			tcnetwork.WithNetworkName([]string{pgAlias}, testNet),
			// The module's default wait strategy dials the mapped *host*
			// port, which this test binary (itself running inside a
			// container on testNet, see network_helper_test.go) has no
			// route to. Skip that external check and rely on the log line
			// instead — postgres logs "ready to accept connections" twice
			// on first boot (once pre-init, once post-init).
			testcontainers.WithAdditionalWaitStrategy(
				wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
				wait.ForListeningPort("5432/tcp").SkipExternalCheck(),
			),
		)
	} else {
		pgOpts = append(pgOpts, postgres.BasicWaitStrategies())
	}

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine", pgOpts...)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	})

	var dsn string
	if testNet != "" {
		dsn = fmt.Sprintf("postgres://postgres:postgres@%s:5432/pokemon?sslmode=disable", pgAlias)
	} else {
		dsn, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			t.Fatalf("failed to get connection string: %v", err)
		}
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, string(schema)); err != nil {
		t.Fatalf("failed to apply schema: %v", err)
	}

	repo := repository.NewPostgresLookupLogRepository(pool)

	fetchedAt := time.Now().UTC().Truncate(time.Second)
	if err := repo.LogLookup(ctx, "mewtwo", fetchedAt); err != nil {
		t.Fatalf("LogLookup failed: %v", err)
	}

	var name string
	var storedAt time.Time
	row := pool.QueryRow(ctx, "SELECT name, fetched_at FROM pokemon_lookup_log WHERE name = $1", "mewtwo")
	if err := row.Scan(&name, &storedAt); err != nil {
		t.Fatalf("failed to query inserted row: %v", err)
	}

	if name != "mewtwo" {
		t.Errorf("expected name %q, got %q", "mewtwo", name)
	}
	if !storedAt.Equal(fetchedAt) {
		t.Errorf("expected fetched_at %v, got %v", fetchedAt, storedAt)
	}
}
