// Package cache provides a Redis-backed cache for PokéAPI lookups, used as
// the cache dependency.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"pokemon-cache-service/internal/pokeapi"
)

//go:generate mockgen -source=pokemon_cache.go -destination=../mock/cache/pokemon_cache_mock.go -package=mockcache

// PokemonCache caches PokemonData by pokemon name.
type PokemonCache interface {
	Get(ctx context.Context, name string) (*pokeapi.PokemonData, bool, error)
	Set(ctx context.Context, name string, data *pokeapi.PokemonData, ttl time.Duration) error
}

type redisCache struct {
	client *redis.Client
}

// NewRedisCache builds a PokemonCache backed by the given go-redis client.
func NewRedisCache(client *redis.Client) PokemonCache {
	return &redisCache{client: client}
}

func cacheKey(name string) string {
	return fmt.Sprintf("pokemon:%s", name)
}

func (c *redisCache) Get(ctx context.Context, name string) (*pokeapi.PokemonData, bool, error) {
	raw, err := c.client.Get(ctx, cacheKey(name)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("getting %q from cache: %w", name, err)
	}

	var data pokeapi.PokemonData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, false, fmt.Errorf("unmarshaling cached value for %q: %w", name, err)
	}

	return &data, true, nil
}

func (c *redisCache) Set(ctx context.Context, name string, data *pokeapi.PokemonData, ttl time.Duration) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling value for %q: %w", name, err)
	}

	if err := c.client.Set(ctx, cacheKey(name), raw, ttl).Err(); err != nil {
		return fmt.Errorf("setting %q in cache: %w", name, err)
	}

	return nil
}
