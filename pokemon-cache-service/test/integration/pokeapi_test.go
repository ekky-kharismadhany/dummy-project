//go:build integration

package integration

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"pokemon-cache-service/internal/pokeapi"
)

// TestPokeAPIClient_AgainstStubServer points the real pokeapi.Client at a
// local httptest.Server returning fixed PokéAPI-shaped JSON, avoiding any
// dependency on the real, rate-limited public API in CI.
func TestPokeAPIClient_AgainstStubServer(t *testing.T) {
	fixture, err := os.ReadFile("../testdata/pokemon_fixture.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/pokemon/eevee":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(fixture)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer stub.Close()

	client := pokeapi.NewClient(stub.URL)

	data, err := client.GetPokemon(context.Background(), "eevee")
	if err != nil {
		t.Fatalf("GetPokemon failed: %v", err)
	}

	if data.Name != "eevee" || data.Height != 3 || data.Weight != 65 || data.BaseExperience != 65 {
		t.Errorf("unexpected data: %+v", data)
	}

	_, err = client.GetPokemon(context.Background(), "missingno")
	if err == nil {
		t.Fatal("expected an error for an unknown pokemon")
	}
	var notFound *pokeapi.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %v", err)
	}
}
