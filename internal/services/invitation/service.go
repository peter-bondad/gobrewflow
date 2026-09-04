package invitation

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"gobrewflow/internal/config"
	"gobrewflow/internal/services/account"
	"gobrewflow/internal/services/user"
	"gobrewflow/internal/utils"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type InvitationService interface {
	SendInvitation(ctx context.Context, email string, inviterID uuid.UUID) (*Invitation, error)
	AcceptInvitation(ctx context.Context, invitationToken string) (*Invitation, error)
	SetPassword(ctx context.Context, input SetPasswordInput) (*Invitation, error)
	CancelInvitation(ctx context.Context, id uuid.UUID, requesterID uuid.UUID) error
	GetInvitation(ctx context.Context, id uuid.UUID, requesterID uuid.UUID) (*Invitation, error)
	ListInvitations(ctx context.Context, inviterID uuid.UUID) ([]Invitation, error)
}

type invitationService struct {
	db               *bun.DB
	invitationRepo   InvitationRepository
	userRepo         user.UserRepository
	accountRepo      account.AccountRepository
	invitationConfig config.InvitationConfig
}

func NewInvitationService(db *bun.DB, invitationRepo InvitationRepository, userRepo user.UserRepository, accountRepo account.AccountRepository, invitationConfig config.InvitationConfig) InvitationService {
	return &invitationService{
		db:               db,
		invitationRepo:   invitationRepo,
		userRepo:         userRepo,
		accountRepo:      accountRepo,
		invitationConfig: invitationConfig,
	}
}

func (s *invitationService) SendInvitation(ctx context.Context, email string, inviterID uuid.UUID) (*Invitation, error) {
	// Validate inviter exists
	inviter, err := s.userRepo.FindByID(ctx, inviterID)
	if err != nil {
		return nil, ErrInvitationNotFound
	}

	if inviter.Role != user.Owner && inviter.Role != user.Manager {
		return nil, ErrForbidden
	}

	// Check for existing pending invitation
	existing, err := s.invitationRepo.GetPendingInvitationByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, ErrInvitationAlreadySent
	}

	// Check inviter limits (example: max 5 pending invitations)
	count, err := s.invitationRepo.CountPendingInvitations(ctx, inviterID)
	if err != nil {
		return nil, err
	}
	if count >= 5 {
		return nil, ErrInvitationLimitReached
	}

	// Generate secure random token
	invitationToken, err := generateToken()
	if err != nil {
		return nil, err
	}

	// hash generated invitation token for security purpose
	invitationTokenHash := hashToken(invitationToken)

	// TTL is fixed
	expiresAt := time.Now().Add(s.invitationConfig.TTL)

	invitation, err := s.invitationRepo.CreateInvitation(ctx, email, inviterID, invitationTokenHash, expiresAt)
	if err != nil {
		return nil, err
	}

	invitation.InvitationToken = invitationToken
	return invitation, nil
}

func (s *invitationService) AcceptInvitation(
	ctx context.Context,
	invitationToken string,
) (*Invitation, error) {
	var invitation *Invitation

	err := s.db.RunInTx(ctx, nil, func(
		ctx context.Context,
		tx bun.Tx,
	) error {
		var err error

		var invitationTokenHash = hashToken(invitationToken)
		invitation, err = s.invitationRepo.GetInvitationByTokenHash(
			ctx,
			tx,
			invitationTokenHash,
		)
		if err != nil {
			return ErrInvitationNotFound
		}

		if invitation.Status != InvitationPending {
			return ErrInvitationNotPending
		}

		now := time.Now()

		if now.After(invitation.ExpiresAt) {
			return ErrInvitationExpired
		}

		setupToken, err := generateToken()
		if err != nil {
			return err
		}
		setupTokenHash := hashToken(setupToken)
		setupTokenExpiresAt := now.Add(15 * time.Minute)

		if err := s.invitationRepo.AcceptInvitation(
			ctx,
			tx,
			invitation.ID,
			now,
			setupTokenHash,
			setupTokenExpiresAt,
		); err != nil {
			return err
		}

		invitation.Status = InvitationAccepted
		invitation.AcceptedAt = &now
		invitation.SetupTokenHash = &setupTokenHash
		invitation.SetupTokenExpiresAt = &setupTokenExpiresAt
		invitation.SetupToken = setupToken

		return nil
	})

	if err != nil {
		return nil, err
	}

	return invitation, nil
}

type SetPasswordInput struct {
	SetupToken string
	FirstName  string
	LastName   string
	Password   string
}

func (s *invitationService) SetPassword(
	ctx context.Context,
	input SetPasswordInput,
) (*Invitation, error) {

	setupTokenHash := hashToken(input.SetupToken)

	var invitation *Invitation

	err := s.db.RunInTx(ctx, nil, func(
		ctx context.Context,
		tx bun.Tx,
	) error {

		// 1. Find invitation
		var err error

		invitation, err = s.invitationRepo.GetInvitationBySetupTokenHash(
			ctx,
			tx,
			setupTokenHash,
		)
		if err != nil {
			return ErrSetupTokenInvalid
		}

		// 2. Check invitation state
		if invitation.Status != InvitationAccepted {
			return ErrInvitationNotAccepted
		}

		// 3. Check token expiration
		now := time.Now()

		if invitation.SetupTokenExpiresAt == nil ||
			now.After(*invitation.SetupTokenExpiresAt) {
			return ErrSetupTokenExpired
		}

		// 4. Hash password
		passwordHash, err := utils.HashPassword(input.Password)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}

		// 5. Create user
		newUser := &user.User{
			ID:        uuid.New(),
			Email:     invitation.Email,
			FirstName: input.FirstName,
			LastName:  input.LastName,
			FullName:  input.FirstName + " " + input.LastName,
			Role:      user.Staff,
		}

		if err := s.userRepo.InsertUser(
			ctx,
			tx,
			newUser,
		); err != nil {
			if IsUniqueViolation(err) {
				return ErrEmailAlreadyExists
			}
			return fmt.Errorf("create user: %w", err)
		}

		// 6. Create credentials account
		newAccount := &account.Account{
			ID:                uuid.New(),
			UserID:            newUser.ID,
			Provider:          "credentials",
			ProviderAccountID: newUser.Email,
			PasswordHash:      &passwordHash,
		}

		if err := s.accountRepo.InsertAccount(
			ctx,
			tx,
			newAccount,
		); err != nil {
			return fmt.Errorf("create account: %w", err)
		}

		// 7. Complete invitation
		if err := s.invitationRepo.CompleteInvitation(
			ctx,
			tx,
			invitation.ID,
			newUser.ID,
			now,
		); err != nil {
			return fmt.Errorf("complete invitation: %w", err)
		}

		// Update returned object
		invitation.UserID = &newUser.ID
		invitation.Status = InvitationCompleted

		return nil
	})

	if err != nil {
		return nil, err
	}

	return invitation, nil
}

func (s *invitationService) CancelInvitation(ctx context.Context, id uuid.UUID, requesterID uuid.UUID) error {
	invitation, err := s.invitationRepo.GetInvitationByID(ctx, id)
	if err != nil {
		return ErrInvitationNotFound
	}

	if invitation.InvitedBy != requesterID {
		return ErrForbidden
	}

	if invitation.Status != InvitationPending {
		return ErrInvitationNotPending
	}

	now := time.Now()

	cancelled, err := s.invitationRepo.CancelInvitation(ctx, s.db, invitation.ID, now)

	if err != nil {
		return err
	}

	if !cancelled {
		return ErrInvitationNotPending
	}

	return nil
}

func (s *invitationService) GetInvitation(ctx context.Context, id uuid.UUID, requesterID uuid.UUID) (*Invitation, error) {
	invitation, err := s.invitationRepo.GetInvitationByID(ctx, id)
	if err != nil {
		return nil, ErrInvitationNotFound
	}

	// Only the inviter or the invited user can view
	if invitation.InvitedBy != requesterID && (invitation.UserID == nil || *invitation.UserID != requesterID) {
		return nil, ErrForbidden
	}

	return invitation, nil
}

func (s *invitationService) ListInvitations(ctx context.Context, inviterID uuid.UUID) ([]Invitation, error) {
	return s.invitationRepo.GetInvitationsByInviter(ctx, inviterID)
}

// to generate token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// to hash Invitation Token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return hex.EncodeToString(hash[:])
}
