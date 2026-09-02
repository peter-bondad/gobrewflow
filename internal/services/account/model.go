package account

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Account struct {
	bun.BaseModel `bun:"table:accounts"`

	ID uuid.UUID `bun:",pk"`

	UserID            uuid.UUID  `bun:"user_id,notnull"`
	Provider          string     `bun:"provider,notnull"`
	ProviderAccountID string     `bun:"provider_account_id,notnull"`
	PasswordHash      *string    `bun:"password_hash"`
	AccessToken       *string    `bun:"access_token"`
	RefreshToken      *string    `bun:"refresh_token"`
	ExpiresAt         *time.Time `bun:"expires_at"`

	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}
