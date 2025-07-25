package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/AvalosM/short-url-service/internal/cache"
	"github.com/AvalosM/short-url-service/internal/config"
	"github.com/AvalosM/short-url-service/internal/handlers"
	"github.com/AvalosM/short-url-service/internal/router"
	"github.com/AvalosM/short-url-service/internal/storage"
	"github.com/AvalosM/short-url-service/pkg/logging"
	"github.com/AvalosM/short-url-service/pkg/metrics"
	"github.com/AvalosM/short-url-service/pkg/shorturl"
)

func main() {
	// TODO: Read configuration from environment variables or a config file
	cfg := config.DefaultConfig()
	err := cfg.Validate()
	if err != nil {
		os.Exit(-1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(cfg.Logger.Level),
	}))

	shutdownOnError := func(err error) {
		if err != nil {
			logger.Error("error initializing application, shutting down", logging.ErrorKey, err)
			os.Exit(-1)
		}
	}

	storage, err := storage.NewStorage(cfg.Storage)
	shutdownOnError(err)

	cache := cache.NewCache(cfg.Cache)

	metricsManager, err := metrics.NewManager(cfg.MetricsManager, storage, logger)
	shutdownOnError(err)

	stopMetricsManager := metricsManager.Start()
	defer stopMetricsManager()

	shortURLManager, err := shorturl.NewManager(cfg.ShortURLManager, storage, cache, logger)
	shutdownOnError(err)

	shortURLHandler, err := handlers.NewShortURLHandler(shortURLManager, metricsManager, logger)
	shutdownOnError(err)

	router := router.NewRouter(cfg.Router, shortURLHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", cfg.HTTPServer.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("Starting server on port", "port", cfg.HTTPServer.Port)
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		shutdownOnError(err)
	}

	os.Exit(0)
}
