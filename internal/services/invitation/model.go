package invitation

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type InvitationStatus string

const (
	InvitationPending   InvitationStatus = "pending"
	InvitationAccepted  InvitationStatus = "accepted"
	InvitationCompleted InvitationStatus = "completed"
	InvitationExpired   InvitationStatus = "expired"
	InvitationCancelled InvitationStatus = "cancelled"
	InvitationRevoked   InvitationStatus = "revoked"
)

type Invitation struct {
	bun.BaseModel `bun:"table:invitations,alias:inv"`

	ID uuid.UUID `bun:",pk"`

	Email       string           `bun:"email,notnull"`
	Token       string           `bun:"token,notnull"`
	InvitedBy   uuid.UUID        `bun:"invited_by,notnull"`
	UserID      *uuid.UUID       `bun:"user_id"`
	Status      InvitationStatus `bun:"status,notnull,default:'pending'"`
	ExpiresAt   time.Time        `bun:"expires_at,notnull"`
	AcceptedAt  *time.Time       `bun:"accepted_at"`
	CancelledAt *time.Time       `bun:"cancelled_at"`

	SetupTokenHash      *string    `bun:"setup_token_hash"`
	SetupTokenExpiresAt *time.Time `bun:"setup_token_expires_at"`
	SetupToken          string     `bun:"-"`

	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}
