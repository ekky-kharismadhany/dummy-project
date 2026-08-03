// Package handler exposes the service's HTTP API.
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"pokemon-cache-service/internal/pokeapi"
	"pokemon-cache-service/internal/service"
)

type errorResponse struct {
	Error string `json:"error"`
}

// NewMux builds an http.ServeMux exposing GET /pokemon/{name}.
func NewMux(svc *service.PokemonService, logger *slog.Logger) *http.ServeMux {
	if logger == nil {
		logger = slog.Default()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /pokemon/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "missing pokemon name"})
			return
		}

		result, err := svc.GetPokemon(r.Context(), name)
		if err != nil {
			var notFound *pokeapi.NotFoundError
			if errors.As(err, &notFound) {
				writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
				return
			}
			logger.Error("pokemon lookup failed", "name", name, "error", err)
			writeJSON(w, http.StatusBadGateway, errorResponse{Error: "failed to fetch pokemon data"})
			return
		}

		writeJSON(w, http.StatusOK, result)
	})

	return mux
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
