package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"taskflow/backend/internal/config"
	"taskflow/backend/internal/database"
	"taskflow/backend/internal/httpapi"
	"taskflow/backend/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	if err := database.RunMigrations(cfg.DatabaseURL, cfg.MigrationDir); err != nil {
		logger.Error("run migrations", "error", err)
		os.Exit(1)
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.ConnectPool(rootCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	dataStore := store.New(pool)
	handler := httpapi.NewHandler(dataStore, logger, cfg.JWTSecret, cfg.JWTExpiry, cfg.BcryptCost)
	router := httpapi.NewRouter(handler, logger, cfg.JWTSecret, cfg.AllowedOrigin)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting api server", "port", cfg.Port)
		if serveErr := srv.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
		}
	}()

	select {
	case <-rootCtx.Done():
		logger.Info("shutdown signal received")
	case serveErr := <-errCh:
		logger.Error("server failed", "error", serveErr)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}

	logger.Info("server stopped")
}
