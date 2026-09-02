package server

import (
	"gobrewflow/internal/middleware"
	"gobrewflow/internal/services/invitation"
	"gobrewflow/internal/services/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func (s *Server) routes() {
	s.Server.GET("/health", s.handleHealth)
}

func (s *Server) publicRoutes(userHandler user.UserHandler, invitationHandler invitation.InvitationHandler) {

	api := s.Server.Group("/api")
	api.POST("/login", userHandler.Login)
	api.POST("/logout", userHandler.Logout)

	// invitation for user routes
	api.POST("/accept-invitation", invitationHandler.AcceptInvitation)
	api.POST("/set-password", invitationHandler.SetPassword)
}

// internal/server/routes.go
func (s *Server) protectedRoutes(
	authMiddleware *middleware.AuthMiddleware,
	userRepo user.UserRepository,
	invitationHandler invitation.InvitationHandler,
) {
	protectedAPI := s.Server.Group("/api/protected")

	protectedAPI.Use(authMiddleware.AuthHandler())

	invitationAPI := protectedAPI.Group("/invitations")

	invitationAPI.Use(
		middleware.RequireRoles(
			userRepo,
			user.Owner,
			user.Manager,
		),
	)

	invitationAPI.POST("/send", invitationHandler.SendInvitation)
	invitationAPI.POST(":id/cancel", invitationHandler.CancelInvitation)
	invitationAPI.GET("/", invitationHandler.ListInvitations)
}
