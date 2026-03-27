package main

import (
	"os"

	"github.com/kleffio/platform/internal/bootstrap"
	"github.com/kleffio/platform/internal/shared/logging"
)

func main() {
	// Initialise the logger early using LOG_LEVEL so that bootstrap messages
	// respect the configured level. The level is read directly from the
	// environment here; bootstrap.LoadConfig validates it again later.
	logger := logging.NewLogger(os.Getenv("LOG_LEVEL"))

	app, err := bootstrap.NewApp(logger)
	if err != nil {
		logger.Error("failed to initialise application", "error", err)
		os.Exit(1)
	}

	if err := app.Run(); err != nil {
		logger.Error("application exited with error", "error", err)
		os.Exit(1)
	}
}
