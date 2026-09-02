package middleware

import (
	"net/http"
	"strings"

	"gobrewflow/internal/services/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	JWTService         *auth.JWTService
	TokenBlacklistRepo auth.TokenBlacklistRepository
}

func NewAuthMiddleware(jwtService *auth.JWTService, tokenBlacklistRepo auth.TokenBlacklistRepository) *AuthMiddleware {
	return &AuthMiddleware{
		JWTService:         jwtService,
		TokenBlacklistRepo: tokenBlacklistRepo,
	}
}

func (m *AuthMiddleware) AuthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// check if auth header is empty string
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header missing",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization format",
			})
			return
		}

		// parse jwt token
		claims, err := m.JWTService.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}

		// convert extracted user id from jwt claims to string
		userIDString, ok := claims["user_id"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "user id missing",
			})
			return
		}

		// then convert to uuid
		// user id in databaes is type uuid so need to convert
		userID, err := uuid.Parse(userIDString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user id",
			})
			return
		}

		jti, ok := m.JWTService.ExtractJTI(claims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token claims",
			})
			return
		}

		revoked, err := m.TokenBlacklistRepo.IsTokenRevoked(c.Request.Context(), jti)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "failed to check token status",
			})
			return
		}
		if revoked {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "token has been revoked",
			})
			return
		}

		c.Set(UserIDKey, userID)

		c.Next()
	}
}
