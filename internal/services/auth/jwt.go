package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	Secret []byte
}

func (j *JWTService) GenerateToken(userID string) (string, string, error) {
	now := time.Now()

	jtiBytes := make([]byte, 16)
	if _, err := rand.Read(jtiBytes); err != nil {
		return "", "", err
	}
	jti := hex.EncodeToString(jtiBytes)

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     now.Add(24 * time.Hour).Unix(),
		"iat":     now.Unix(),
		"jti":     jti,
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signed, err := token.SignedString(j.Secret)
	if err != nil {
		return "", "", err
	}

	return signed, jti, nil
}

func (j *JWTService) ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method")
			}

			return j.Secret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	return claims, nil
}

func (j *JWTService) ExtractJTI(claims jwt.MapClaims) (string, bool) {
	jti, ok := claims["jti"].(string)
	return jti, ok
}

func (j *JWTService) ExtractExpiration(claims jwt.MapClaims) (string, bool) {
	exp, ok := claims["exp"].(string)
	return exp, ok
}
