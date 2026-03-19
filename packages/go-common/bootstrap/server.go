package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const defaultShutdownTimeout = 15 * time.Second

// ServerConfig holds the options for RunServer.
type ServerConfig struct {
	// Port the HTTP server listens on (default: 8080).
	Port int
	// Handler is the root http.Handler (typically a mux with all routes attached).
	Handler http.Handler
	// Logger is used for startup, shutdown, and error messages.
	Logger *slog.Logger
	// ShutdownTimeout is how long to wait for in-flight requests to finish.
	// Defaults to 15 seconds.
	ShutdownTimeout time.Duration
}

// RunServer starts the HTTP server and blocks until SIGINT or SIGTERM is
// received, then performs a graceful shutdown.
//
// Usage:
//
//	if err := bootstrap.RunServer(bootstrap.ServerConfig{
//	    Port:    8080,
//	    Handler: mux,
//	    Logger:  logger,
//	}); err != nil {
//	    os.Exit(1)
//	}
func RunServer(cfg ServerConfig) error {
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = defaultShutdownTimeout
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      cfg.Handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background.
	errCh := make(chan error, 1)
	go func() {
		cfg.Logger.Info("server listening", slog.Int("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	// Wait for shutdown signal or server error.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		cfg.Logger.Info("shutdown signal received", slog.String("signal", sig.String()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	cfg.Logger.Info("server stopped gracefully")
	return nil
}
