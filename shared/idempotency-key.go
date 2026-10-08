package shared

import "github.com/google/uuid"

func GenerateIdempotencyKey() string {
	return uuid.New().String()
}
