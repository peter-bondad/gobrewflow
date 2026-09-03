package user

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	Login(c *gin.Context)
	Logout(c *gin.Context)
	ListUsers(c *gin.Context)
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

type ListUsersRequest struct {
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
	Role   string `form:"role"`
}

func (h *userHandler) ListUsers(c *gin.Context) {
	var req ListUsersRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid query parameters",
		})
		return
	}

	input := UserListInput{
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	if req.Role != "" {
		role := UserRole(req.Role)
		input.UserRole = &role
	}

	users, err := h.service.ListUsers(
		c.Request.Context(),
		input,
	)
	if err != nil {
		switch err {
		case ErrInvalidUserRole:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid user role",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to list users",
			})
		}
		return
	}

	c.JSON(http.StatusOK, users)
}
