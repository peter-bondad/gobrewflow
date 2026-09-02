package user

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	Login(c *gin.Context)
	Logout(c *gin.Context)
}

type userHandler struct {
	service UserService
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func NewUserHandler(service UserService) UserHandler {
	return &userHandler{
		service: service,
	}
}
func (h *userHandler) Login(c *gin.Context) {
	var input LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(err)
		return
	}

	response, err := h.service.Login(
		c.Request.Context(),
		input,
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *userHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header missing"})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
		return
	}

	if err := h.service.Logout(c.Request.Context(), parts[1]); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	c.Status(http.StatusNoContent)
}
