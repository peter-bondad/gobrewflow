package app

import (
	"context"
	"gobrewflow/internal/config"
	"gobrewflow/internal/server"
	"log/slog"

	"github.com/uptrace/bun"
)

type App struct {
	cfg    *config.Config
	log    *slog.Logger
	server *server.Server
}

// New creates a new instance of the App struct, initializing the server and other components based on the provided configuration and logger.
func New(cfg *config.Config, log *slog.Logger, db *bun.DB) (*App, error) {
	container := NewContainer(db, cfg)

	srv, err := server.New(cfg, log, server.Dependencies{
		UserRepo:           container.UserRepo,
		UserHandler:        container.UserHandler,
		InvitationHandler:  container.InvitationHandler,
		JwtService:         container.JwtService,
		TokenBlacklistRepo: container.TokenBlacklistRepo,
	})
	if err != nil {
		return nil, err
	}

	app := &App{
		cfg:    cfg,
		log:    log,
		server: srv,
	}

	return app, nil
}

// Run starts the application by starting the server. It blocks until the server is stopped or an error occurs.
func (a *App) Run() error {
	return a.server.Start(a.cfg)
}

// Shutdown attempts to gracefully shut down the application by shutting down the server.
func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
