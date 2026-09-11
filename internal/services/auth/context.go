package auth

import (
	"errors"
	"gobrewflow/shared"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

var (
	ErrUnauthorized  = errors.New("unauthorized")
	ErrInvalidUserID = errors.New("invalid user ID")
)

func GetUserID(c *gin.Context) (uuid.UUID, error) {
	value, exists := c.Get(shared.UserIDKey)
	if !exists {
		return uuid.Nil, ErrUnauthorized
	}

	userID, ok := value.(uuid.UUID)
	if !ok {
		return uuid.Nil, ErrInvalidUserID
	}

	return userID, nil
}
