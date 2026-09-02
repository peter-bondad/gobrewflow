package middleware

import (
	"context"
	"net/http"
	"slices"

	"gobrewflow/internal/services/user"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserFinder interface {
	FindByID(ctx context.Context, id uuid.UUID) (*user.User, error)
}

func RequireRoles(
	finder UserFinder,
	allowed ...user.UserRole,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "unauthorized"},
			)
			return
		}

		u, err := finder.FindByID(
			c.Request.Context(),
			userID,
		)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"error": "failed to load user"},
			)
			return
		}

		if slices.Contains(allowed, u.Role) {
			c.Set("user", u)
			c.Next()
			return
		}

		c.AbortWithStatusJSON(
			http.StatusForbidden,
			gin.H{"error": "insufficient permissions"},
		)
	}
}
