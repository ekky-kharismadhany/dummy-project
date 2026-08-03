// Command server runs the pokemon-cache-service HTTP API.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"pokemon-cache-service/internal/cache"
	"pokemon-cache-service/internal/handler"
	"pokemon-cache-service/internal/pokeapi"
	"pokemon-cache-service/internal/repository"
	"pokemon-cache-service/internal/service"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	postgresDSN := getenv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/pokemon?sslmode=disable")
	redisAddr := getenv("REDIS_ADDR", "localhost:6379")
	pokeAPIBaseURL := getenv("POKEAPI_BASE_URL", "https://pokeapi.co/api/v2")
	listenAddr := getenv("LISTEN_ADDR", ":8080")

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, postgresDSN)
	if err != nil {
		logger.Error("failed to create postgres pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer redisClient.Close()

	repo := repository.NewPostgresLookupLogRepository(pool)
	pokemonCache := cache.NewRedisCache(redisClient)
	pokeAPIClient := pokeapi.NewClient(pokeAPIBaseURL)

	svc := service.NewPokemonService(pokemonCache, repo, pokeAPIClient, logger)
	mux := handler.NewMux(svc, logger)

	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("starting server", "addr", listenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("shutting down")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
	}
}
