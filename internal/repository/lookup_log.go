// Package repository provides the Postgres-backed lookup log, used as the
// database dependency.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:generate mockgen -source=lookup_log.go -destination=../mock/repository/lookup_log_mock.go -package=mockrepository

// LookupLogRepository records every Pokémon lookup performed by the service.
type LookupLogRepository interface {
	LogLookup(ctx context.Context, name string, fetchedAt time.Time) error
}

type postgresLookupLogRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresLookupLogRepository builds a LookupLogRepository backed by the
// given pgx pool.
func NewPostgresLookupLogRepository(pool *pgxpool.Pool) LookupLogRepository {
	return &postgresLookupLogRepository{pool: pool}
}

func (r *postgresLookupLogRepository) LogLookup(ctx context.Context, name string, fetchedAt time.Time) error {
	const query = `INSERT INTO pokemon_lookup_log (name, fetched_at) VALUES ($1, $2)`

	if _, err := r.pool.Exec(ctx, query, name, fetchedAt); err != nil {
		return fmt.Errorf("inserting lookup log for %q: %w", name, err)
	}

	return nil
}
