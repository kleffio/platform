package bootstrap

import (
	"fmt"
	"log/slog"

	"github.com/kleff/go-common/bootstrap"
)

// App is the top-level application object. It owns the lifecycle of all
// components: config loading, dependency wiring, HTTP server, and shutdown.
type App struct {
	config    *Config
	container *Container
	logger    *slog.Logger
}

// NewApp initialises the application. Returns an error if config is invalid
// or a dependency cannot be wired up.
func NewApp(logger *slog.Logger) (*App, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	container, err := NewContainer(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("build container: %w", err)
	}

	return &App{
		config:    cfg,
		container: container,
		logger:    logger,
	}, nil
}

// Run starts the HTTP server and blocks until a shutdown signal is received.
func (a *App) Run() error {
	handler := buildRouter(a.container)

	return bootstrap.RunServer(bootstrap.ServerConfig{
		Port:    a.config.HTTPPort,
		Handler: handler,
		Logger:  a.logger,
	})
}
