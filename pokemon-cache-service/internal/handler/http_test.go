package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/mock/gomock"

	"pokemon-cache-service/internal/handler"
	mockcache "pokemon-cache-service/internal/mock/cache"
	mockpokeapi "pokemon-cache-service/internal/mock/pokeapi"
	mockrepository "pokemon-cache-service/internal/mock/repository"
	"pokemon-cache-service/internal/pokeapi"
	"pokemon-cache-service/internal/service"
)

func TestPokemonEndpoint_ReturnsAPIResult(t *testing.T) {
	ctrl := gomock.NewController(t)

	c := mockcache.NewMockPokemonCache(ctrl)
	r := mockrepository.NewMockLookupLogRepository(ctrl)
	api := mockpokeapi.NewMockClient(ctrl)

	fetched := &pokeapi.PokemonData{Name: "ditto", Height: 1, Weight: 2, BaseExperience: 3}
	c.EXPECT().Get(gomock.Any(), "ditto").Return(nil, false, nil)
	api.EXPECT().GetPokemon(gomock.Any(), "ditto").Return(fetched, nil)
	c.EXPECT().Set(gomock.Any(), "ditto", fetched, service.CacheTTL).Return(nil)
	r.EXPECT().LogLookup(gomock.Any(), "ditto", gomock.Any()).Return(nil)

	svc := service.NewPokemonService(c, r, api, nil)
	mux := handler.NewMux(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/pokemon/ditto", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var result service.Result
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Name != "ditto" || result.Source != service.SourceAPI {
		t.Errorf("unexpected result: %+v", result)
	}
}
