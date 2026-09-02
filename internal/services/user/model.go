package user

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type UserRole string

const (
	Owner   UserRole = "owner"
	Manager UserRole = "manager"
	Staff   UserRole = "staff"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID uuid.UUID `bun:",pk"`

	Email string

	FullName string `bun:"full_name"`

	FirstName  string  `bun:"first_name"`
	MiddleName *string `bun:"middle_name"`
	LastName   string  `bun:"last_name"`

	PhoneNumber *string

	Role UserRole `bun:"role"`

	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}
