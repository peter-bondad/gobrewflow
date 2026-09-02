package invitation

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type InvitationRepository interface {
	CreateInvitation(ctx context.Context, email string, inviterID uuid.UUID, token string, expiresAt time.Time) (*Invitation, error)
	GetInvitationByToken(ctx context.Context, db bun.IDB, token string) (*Invitation, error)
	GetInvitationByID(ctx context.Context, id uuid.UUID) (*Invitation, error)
	GetPendingInvitationByEmail(ctx context.Context, email string) (*Invitation, error)
	GetInvitationsByInviter(ctx context.Context, inviterID uuid.UUID) ([]Invitation, error)
	CountPendingInvitations(ctx context.Context, inviterID uuid.UUID) (int, error)
	AcceptInvitation(ctx context.Context, db bun.IDB, invitationID uuid.UUID, now time.Time, setupTokenHash string, setupTokenHashExpiresAt time.Time) error
	CancelInvitation(ctx context.Context, db bun.IDB, invitationID uuid.UUID, now time.Time) (bool, error)
	DeleteInvitation(ctx context.Context, id uuid.UUID) error
	GetInvitationBySetupTokenHash(ctx context.Context, db bun.IDB, setupTokenHash string) (*Invitation, error)
	CompleteInvitation(ctx context.Context, db bun.IDB, invitationID uuid.UUID, userID uuid.UUID, now time.Time) error
}

type invitationRepository struct {
	db *bun.DB
}

func NewInvitationRepository(db *bun.DB) InvitationRepository {
	return &invitationRepository{
		db: db,
	}
}

func (r *invitationRepository) CreateInvitation(ctx context.Context, email string, inviterID uuid.UUID, token string, expiresAt time.Time) (*Invitation, error) {
	invitation := &Invitation{
		ID:        uuid.New(),
		Email:     email,
		InvitedBy: inviterID,
		Token:     token,
		ExpiresAt: expiresAt,
		Status:    InvitationPending,
	}

	_, err := r.db.NewInsert().Model(invitation).Returning("*").Exec(ctx, invitation)
	if err != nil {
		return nil, err
	}

	return invitation, nil
}

func (r *invitationRepository) GetInvitationByToken(
	ctx context.Context,
	q bun.IDB,
	token string,
) (*Invitation, error) {
	invitation := new(Invitation)

	err := q.NewSelect().
		Model(invitation).
		Where("token = ?", token).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return invitation, nil
}

func (r *invitationRepository) GetInvitationByID(ctx context.Context, id uuid.UUID) (*Invitation, error) {
	invitation := new(Invitation)
	err := r.db.NewSelect().Model(invitation).Where("id = ?", id).Scan(ctx, invitation)
	if err != nil {
		return nil, err
	}
	return invitation, nil
}

func (r *invitationRepository) GetPendingInvitationByEmail(ctx context.Context, email string) (*Invitation, error) {
	invitation := new(Invitation)
	err := r.db.NewSelect().Model(invitation).
		Where("email = ?", email).
		Where("status = ?", InvitationPending).
		Where("expires_at > ?", time.Now()).
		Scan(ctx, invitation)
	if err != nil {
		return nil, err
	}
	return invitation, nil
}

func (r *invitationRepository) GetInvitationsByInviter(ctx context.Context, inviterID uuid.UUID) ([]Invitation, error) {
	var invitations []Invitation
	err := r.db.NewSelect().Model(&invitations).
		Where("invited_by = ?", inviterID).
		Order("created_at DESC").
		Scan(ctx, &invitations)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

func (r *invitationRepository) CountPendingInvitations(ctx context.Context, inviterID uuid.UUID) (int, error) {
	count, err := r.db.NewSelect().Model((*Invitation)(nil)).
		Where("invited_by = ?", inviterID).
		Where("status = ?", InvitationPending).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *invitationRepository) AcceptInvitation(
	ctx context.Context,
	db bun.IDB,
	invitationID uuid.UUID,
	now time.Time,
	setupTokenHash string,
	setupTokenHashExpiresAt time.Time,
) error {
	result, err := db.NewUpdate().
		Model((*Invitation)(nil)).
		Set("status = ?", InvitationAccepted).
		Set("setup_token_hash = ?", setupTokenHash).
		Set("setup_token_expires_at = ?", setupTokenHashExpiresAt).
		Set("accepted_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", invitationID).
		Where("status = ?", InvitationPending).
		Where("expires_at > ?", now).
		Exec(ctx)

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return ErrInvitationNotPending
	}
	return err
}

func (r *invitationRepository) CancelInvitation(
	ctx context.Context,
	db bun.IDB,
	invitationID uuid.UUID,
	now time.Time,
) (bool, error) {
	result, err := db.NewUpdate().
		Model((*Invitation)(nil)).
		Set("status = ?", InvitationCancelled).
		Set("cancelled_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", invitationID).
		Where("status = ?", InvitationPending).
		Exec(ctx)

	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows == 1, nil
}

func (r *invitationRepository) DeleteInvitation(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().Model((*Invitation)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *invitationRepository) GetInvitationBySetupTokenHash(
	ctx context.Context,
	q bun.IDB,
	setupTokenHash string,
) (*Invitation, error) {
	invitation := new(Invitation)

	err := q.NewSelect().
		Model(invitation).
		Where("setup_token_hash = ?", setupTokenHash).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return invitation, nil
}

func (r *invitationRepository) CompleteInvitation(
	ctx context.Context,
	db bun.IDB,
	invitationID uuid.UUID,
	userID uuid.UUID,
	now time.Time,
) error {
	result, err := db.NewUpdate().
		Model((*Invitation)(nil)).
		Set("status = ?", InvitationCompleted).
		Set("user_id = ?", userID).
		Set("setup_token_hash = NULL").
		Set("setup_token_expires_at = NULL").
		Set("updated_at = ?", now).
		Where("id = ?", invitationID).
		Where("status = ?", InvitationAccepted).
		Exec(ctx)

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrInvitationNotAccepted
	}

	return err
}
