package main

import (
	"context"
	"errors"
	"gobrewflow/internal/app"
	"gobrewflow/internal/config"
	"gobrewflow/internal/database"
	"gobrewflow/shared/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

// main is the entry point of the application. It initializes the configuration, logger, and application components, starts the server, and handles graceful shutdown on receiving termination signals.
func main() {
	// Load environment variables from .env file
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Printf("failed to load config: %v", err)
		os.Exit(1)
	}

	log := logger.New(
		string(cfg.App.Env),
		string(cfg.App.LogLevel),
	)

	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		log.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize the application
	app, err := app.New(cfg, log, db)
	if err != nil {
		log.Error("failed to create application", "error", err)
		os.Exit(1)
	}

	// Run the application in a separate goroutine
	go func() {
		if err := app.Run(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to run application", "error", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		cfg.App.ShutdownTimeout,
	)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to shutdown application", "error", err)
		os.Exit(1)
	}
}
