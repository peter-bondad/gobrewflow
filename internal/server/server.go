package server

import (
	"context"
	"gobrewflow/internal/config"
	"gobrewflow/internal/middleware"
	"gobrewflow/internal/services/auth"
	"gobrewflow/internal/services/invitation"
	"gobrewflow/internal/services/user"
	"gobrewflow/shared/logger"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Server *gin.Engine

	Log *slog.Logger

	HttpServer *http.Server
}

type Dependencies struct {
	UserRepo          user.UserRepository
	UserHandler       user.UserHandler
	InvitationHandler invitation.InvitationHandler
	JwtService        *auth.JWTService
	TokenBlacklistRepo auth.TokenBlacklistRepository
}

func New(cfg *config.Config, log *slog.Logger, deps Dependencies) (*Server, error) {

	router := gin.New()

	router.Use(logger.RequestLogger(log))

	jwtService := &auth.JWTService{
		Secret: []byte(cfg.JWT.Secret),
	}

	authMiddleware := middleware.NewAuthMiddleware(jwtService, deps.TokenBlacklistRepo)

	s := &Server{
		Server:     router,
		Log:        log,
		HttpServer: &http.Server{},
	}

	s.routes()
	s.publicRoutes(deps.UserHandler, deps.InvitationHandler)
	s.protectedRoutes(authMiddleware, deps.UserRepo, deps.InvitationHandler)

	return s, nil
}

// This will be used in main.go to start the server and handle graceful shutdowns
func (s *Server) Start(cfg *config.Config) error {
	s.HttpServer = &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: s.Server,
	}
	s.Log.Info("starting http server", slog.String("port", cfg.App.Port))
	return s.HttpServer.ListenAndServe()
}

// This will be used in main.go to gracefully shut down the server when the application is terminating
func (s *Server) Shutdown(ctx context.Context) error {
	s.Log.Info("shutting down http server")

	return s.HttpServer.Shutdown(ctx)
}
