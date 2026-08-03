// Package pokeapi provides a thin client for the public PokéAPI
// (https://pokeapi.co/docs/v2), used as the third-party API dependency.
package pokeapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PokemonData is the shared DTO returned by the PokéAPI client and stored in
// the cache.
type PokemonData struct {
	Name           string `json:"name"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	BaseExperience int    `json:"base_experience"`
}

//go:generate mockgen -source=client.go -destination=../mock/pokeapi/client_mock.go -package=mockpokeapi

// Client fetches Pokémon data from PokéAPI (or a compatible stub).
type Client interface {
	GetPokemon(ctx context.Context, name string) (*PokemonData, error)
}

// NotFoundError indicates PokéAPI returned a 404 for the requested name.
type NotFoundError struct {
	Name string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("pokemon %q not found", e.Name)
}

type httpClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient builds a Client that talks to the given base URL, e.g.
// "https://pokeapi.co/api/v2" or a local stub server's URL.
func NewClient(baseURL string) Client {
	return &httpClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type pokeAPIResponse struct {
	Name           string `json:"name"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	BaseExperience int    `json:"base_experience"`
}

func (c *httpClient) GetPokemon(ctx context.Context, name string) (*PokemonData, error) {
	url := fmt.Sprintf("%s/pokemon/%s", c.baseURL, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building pokeapi request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling pokeapi: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{Name: name}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pokeapi returned unexpected status %d", resp.StatusCode)
	}

	var body pokeAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decoding pokeapi response: %w", err)
	}

	return &PokemonData{
		Name:           body.Name,
		Height:         body.Height,
		Weight:         body.Weight,
		BaseExperience: body.BaseExperience,
	}, nil
}
