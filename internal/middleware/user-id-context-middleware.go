package middleware

import (
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.Nil, false
	}

	userID, ok := value.(uuid.UUID)
	return userID, ok
}
