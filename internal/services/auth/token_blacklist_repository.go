package auth

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type RevokedToken struct {
	bun.BaseModel `bun:"table:revoked_tokens"`

	JTI       string    `bun:"jti,notnull,pk"`
	ExpiresAt time.Time `bun:"expires_at,notnull"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

type TokenBlacklistRepository interface {
	RevokeToken(ctx context.Context, jti string, expiresAt time.Time) error
	IsTokenRevoked(ctx context.Context, jti string) (bool, error)
}

type tokenBlacklistRepository struct {
	db *bun.DB
}

func NewTokenBlacklistRepository(db *bun.DB) TokenBlacklistRepository {
	return &tokenBlacklistRepository{db: db}
}

func (r *tokenBlacklistRepository) RevokeToken(ctx context.Context, jti string, expiresAt time.Time) error {
	revokedToken := new(RevokedToken)
	revokedToken.JTI = jti
	revokedToken.ExpiresAt = expiresAt
	_, err := r.db.NewInsert().
		Model(revokedToken).
		On("CONFLICT (jti) DO NOTHING").
		Exec(ctx)
	return err
}

func (r *tokenBlacklistRepository) IsTokenRevoked(ctx context.Context, jti string) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*RevokedToken)(nil)).
		Where("jti = ?", jti).
		Where("expires_at > ?", time.Now()).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
