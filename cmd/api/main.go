package main

import (
	"context"
	"errors"
	"gobrewflow/internal/app"
	"gobrewflow/internal/config"
	"gobrewflow/internal/database"
	"gobrewflow/shared/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

// main is the entry point of the application. It initializes the configuration, logger, and application components, starts the server, and handles graceful shutdown on receiving termination signals.
func main() {

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// Initialize logger
	log := logger.New("development", "debug")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load application configuration", "error", err)
		os.Exit(1)
	}

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

	// Set up signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop() // Ensure that the signal handler is stopped when main exits

	<-ctx.Done() // Wait for a termination signal

	// Create a context with a timeout for the shutdown process
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		cfg.App.ShutdownTimeout,
	)
	defer cancel() // Ensure that the shutdown context is canceled when main exits

	// Attempt to gracefully shut down the application
	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to shutdown application", "error", err)
		os.Exit(1)
	}
}
