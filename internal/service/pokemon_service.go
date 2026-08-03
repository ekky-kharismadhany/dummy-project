// Package service contains the business logic that wires the cache,
// repository, and PokéAPI client together to serve Pokémon lookups.
package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"pokemon-cache-service/internal/cache"
	"pokemon-cache-service/internal/pokeapi"
	"pokemon-cache-service/internal/repository"
)

// CacheTTL is how long a PokéAPI response is cached for.
const CacheTTL = 5 * time.Minute

// Source identifies where a Result came from.
type Source string

const (
	SourceCache Source = "cache"
	SourceAPI   Source = "api"
)

// Result is the outward-facing shape returned by the /pokemon/{name} endpoint.
type Result struct {
	Name           string `json:"name"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	BaseExperience int    `json:"base_experience"`
	Source         Source `json:"source"`
}

// PokemonService implements the main GET /pokemon/{name} use case.
type PokemonService struct {
	cache   cache.PokemonCache
	repo    repository.LookupLogRepository
	pokeAPI pokeapi.Client
	logger  *slog.Logger
}

// NewPokemonService builds a PokemonService from its three dependencies. A
// nil logger falls back to slog.Default().
func NewPokemonService(c cache.PokemonCache, r repository.LookupLogRepository, p pokeapi.Client, logger *slog.Logger) *PokemonService {
	if logger == nil {
		logger = slog.Default()
	}
	return &PokemonService{cache: c, repo: r, pokeAPI: p, logger: logger}
}

// GetPokemon resolves a pokemon lookup: cache first, PokéAPI on miss, with
// best-effort cache population and lookup logging. Cache and repository
// failures are logged but never fail the request.
func (s *PokemonService) GetPokemon(ctx context.Context, name string) (*Result, error) {
	if cached, ok, err := s.cache.Get(ctx, name); err != nil {
		s.logger.Warn("cache get failed", "name", name, "error", err)
	} else if ok {
		return toResult(cached, SourceCache), nil
	}

	data, err := s.pokeAPI.GetPokemon(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("fetching %q from pokeapi: %w", name, err)
	}

	if err := s.cache.Set(ctx, name, data, CacheTTL); err != nil {
		s.logger.Warn("cache set failed", "name", name, "error", err)
	}

	if err := s.repo.LogLookup(ctx, name, time.Now()); err != nil {
		s.logger.Warn("logging lookup failed", "name", name, "error", err)
	}

	return toResult(data, SourceAPI), nil
}

func toResult(data *pokeapi.PokemonData, source Source) *Result {
	return &Result{
		Name:           data.Name,
		Height:         data.Height,
		Weight:         data.Weight,
		BaseExperience: data.BaseExperience,
		Source:         source,
	}
}
