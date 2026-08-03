package service_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	mockcache "pokemon-cache-service/internal/mock/cache"
	mockpokeapi "pokemon-cache-service/internal/mock/pokeapi"
	mockrepository "pokemon-cache-service/internal/mock/repository"
	"pokemon-cache-service/internal/pokeapi"
	"pokemon-cache-service/internal/service"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(discardWriter{}, nil))
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestGetPokemon_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)

	c := mockcache.NewMockPokemonCache(ctrl)
	r := mockrepository.NewMockLookupLogRepository(ctrl)
	api := mockpokeapi.NewMockClient(ctrl)

	cached := &pokeapi.PokemonData{Name: "pikachu", Height: 4, Weight: 60, BaseExperience: 112}
	c.EXPECT().Get(gomock.Any(), "pikachu").Return(cached, true, nil)
	// No Set/LogLookup/GetPokemon expectations: gomock fails the test if any
	// of those are called, since a cache hit must not touch the other two.

	svc := service.NewPokemonService(c, r, api, discardLogger())

	result, err := svc.GetPokemon(context.Background(), "pikachu")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Source != service.SourceCache {
		t.Errorf("expected source cache, got %q", result.Source)
	}
	if result.Name != "pikachu" || result.BaseExperience != 112 {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestGetPokemon_CacheMiss_APISuccess(t *testing.T) {
	ctrl := gomock.NewController(t)

	c := mockcache.NewMockPokemonCache(ctrl)
	r := mockrepository.NewMockLookupLogRepository(ctrl)
	api := mockpokeapi.NewMockClient(ctrl)

	fetched := &pokeapi.PokemonData{Name: "bulbasaur", Height: 7, Weight: 69, BaseExperience: 64}
	c.EXPECT().Get(gomock.Any(), "bulbasaur").Return(nil, false, nil)
	api.EXPECT().GetPokemon(gomock.Any(), "bulbasaur").Return(fetched, nil)
	c.EXPECT().Set(gomock.Any(), "bulbasaur", fetched, service.CacheTTL).Return(nil)
	r.EXPECT().LogLookup(gomock.Any(), "bulbasaur", gomock.Any()).DoAndReturn(
		func(_ context.Context, name string, fetchedAt time.Time) error {
			if fetchedAt.IsZero() {
				t.Error("expected a non-zero fetchedAt timestamp")
			}
			return nil
		})

	svc := service.NewPokemonService(c, r, api, discardLogger())

	result, err := svc.GetPokemon(context.Background(), "bulbasaur")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Source != service.SourceAPI {
		t.Errorf("expected source api, got %q", result.Source)
	}
	if result.Name != "bulbasaur" || result.BaseExperience != 64 {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestGetPokemon_APIFailure(t *testing.T) {
	ctrl := gomock.NewController(t)

	c := mockcache.NewMockPokemonCache(ctrl)
	r := mockrepository.NewMockLookupLogRepository(ctrl)
	api := mockpokeapi.NewMockClient(ctrl)

	c.EXPECT().Get(gomock.Any(), "missingno").Return(nil, false, nil)
	api.EXPECT().GetPokemon(gomock.Any(), "missingno").Return(nil, errors.New("pokeapi unavailable"))
	// No Set/LogLookup expectations: an API failure must short-circuit
	// before either is attempted.

	svc := service.NewPokemonService(c, r, api, discardLogger())

	_, err := svc.GetPokemon(context.Background(), "missingno")
	if err == nil {
		t.Fatal("expected an error when the pokeapi call fails")
	}
}

func TestGetPokemon_RepositoryWriteFailure_DoesNotFailRequest(t *testing.T) {
	ctrl := gomock.NewController(t)

	c := mockcache.NewMockPokemonCache(ctrl)
	r := mockrepository.NewMockLookupLogRepository(ctrl)
	api := mockpokeapi.NewMockClient(ctrl)

	fetched := &pokeapi.PokemonData{Name: "charmander", Height: 6, Weight: 85, BaseExperience: 62}
	c.EXPECT().Get(gomock.Any(), "charmander").Return(nil, false, nil)
	api.EXPECT().GetPokemon(gomock.Any(), "charmander").Return(fetched, nil)
	c.EXPECT().Set(gomock.Any(), "charmander", fetched, service.CacheTTL).Return(nil)
	r.EXPECT().LogLookup(gomock.Any(), "charmander", gomock.Any()).Return(errors.New("db write failed"))

	svc := service.NewPokemonService(c, r, api, discardLogger())

	result, err := svc.GetPokemon(context.Background(), "charmander")
	if err != nil {
		t.Fatalf("expected repository write failure to be swallowed, got error: %v", err)
	}
	if result.Source != service.SourceAPI {
		t.Errorf("expected source api, got %q", result.Source)
	}
}

func TestGetPokemon_CacheGetFailure_FallsBackToAPI(t *testing.T) {
	ctrl := gomock.NewController(t)

	c := mockcache.NewMockPokemonCache(ctrl)
	r := mockrepository.NewMockLookupLogRepository(ctrl)
	api := mockpokeapi.NewMockClient(ctrl)

	fetched := &pokeapi.PokemonData{Name: "squirtle", Height: 5, Weight: 90, BaseExperience: 63}
	c.EXPECT().Get(gomock.Any(), "squirtle").Return(nil, false, errors.New("cache unavailable"))
	api.EXPECT().GetPokemon(gomock.Any(), "squirtle").Return(fetched, nil)
	c.EXPECT().Set(gomock.Any(), "squirtle", fetched, service.CacheTTL).Return(nil)
	r.EXPECT().LogLookup(gomock.Any(), "squirtle", gomock.Any()).Return(nil)

	svc := service.NewPokemonService(c, r, api, discardLogger())

	result, err := svc.GetPokemon(context.Background(), "squirtle")
	if err != nil {
		t.Fatalf("expected cache get failure to be swallowed, got error: %v", err)
	}
	if result.Source != service.SourceAPI {
		t.Errorf("expected source api after cache get failure, got %q", result.Source)
	}
}

func TestGetPokemon_CacheSetFailure_DoesNotFailRequest(t *testing.T) {
	ctrl := gomock.NewController(t)

	c := mockcache.NewMockPokemonCache(ctrl)
	r := mockrepository.NewMockLookupLogRepository(ctrl)
	api := mockpokeapi.NewMockClient(ctrl)

	fetched := &pokeapi.PokemonData{Name: "jigglypuff", Height: 5, Weight: 55, BaseExperience: 76}
	c.EXPECT().Get(gomock.Any(), "jigglypuff").Return(nil, false, nil)
	api.EXPECT().GetPokemon(gomock.Any(), "jigglypuff").Return(fetched, nil)
	c.EXPECT().Set(gomock.Any(), "jigglypuff", fetched, service.CacheTTL).Return(errors.New("cache write failed"))
	r.EXPECT().LogLookup(gomock.Any(), "jigglypuff", gomock.Any()).Return(nil)

	svc := service.NewPokemonService(c, r, api, discardLogger())

	result, err := svc.GetPokemon(context.Background(), "jigglypuff")
	if err != nil {
		t.Fatalf("expected cache set failure to be swallowed, got error: %v", err)
	}
	if result.Source != service.SourceAPI {
		t.Errorf("expected source api, got %q", result.Source)
	}
}
