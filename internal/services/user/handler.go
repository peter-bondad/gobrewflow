package user

import (
	"gobrewflow/shared"
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

func NewUserHandler(service UserService) UserHandler {
	return &userHandler{
		service: service,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (h *userHandler) Login(c *gin.Context) {
	var input LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(err)
		return
	}

	token, err := h.service.Login(
		c.Request.Context(),
		LoginInput{
			Email:    input.Email,
			Password: input.Password,
		},
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
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
	FullName string    `form:"full_name"`
	Email    string    `form:"email"`
	Role     *UserRole `form:"role"`
	Limit    int       `form:"limit" binding:"gte=0,lte=100"`
	Page     int       `form:"page" binding:"gte=1"`
}
type ListUser struct {
	FullName string   `json:"full_name"`
	Email    string   `json:"email"`
	Role     UserRole `json:"role"`
}

type ListUserResponse struct {
	Data       []ListUser        `json:"data"`
	Pagination shared.Pagination `json:"pagination"`
}

func (h *userHandler) ListUsers(c *gin.Context) {
	var req ListUsersRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid query parameters",
		})
		return
	}

	input := ListUserInput{
		FullName: req.FullName,
		Email:    req.Email,
		Limit:    req.Limit,
		Page:     req.Page,
	}

	if req.Role != nil {
		role := UserRole(*req.Role)
		input.Role = &role
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

	data := make([]ListUser, 0, len(users.Data))

	for _, user := range users.Data {
		data = append(data, ListUser{
			FullName: user.FullName,
			Email:    user.Email,
			Role:     user.Role,
		})
	}

	resp := ListUserResponse{
		Data: data,
		Pagination: shared.Pagination{
			Page:       users.Page,
			Limit:      users.Limit,
			Total:      users.Total,
			TotalPages: users.TotalPages,
		},
	}

	c.JSON(http.StatusOK, resp)
}
