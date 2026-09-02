package middleware

import (
	"errors"
	"gobrewflow/internal/services/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		switch {
		case errors.Is(err, user.InvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid credentials",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
		}
	}
}
