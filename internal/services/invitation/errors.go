package invitation

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrInvitationNotFound        = errors.New("invitation not found")
	ErrInvitationAlreadySent     = errors.New("invitation already sent to this email")
	ErrInvitationLimitReached    = errors.New("invitation limit reached")
	ErrInvitationAlreadyAccepted = errors.New("invitation already accepted")
	ErrInvitationNotPending      = errors.New("invitation is not pending")
	ErrInvitationExpired         = errors.New("invitation has expired")
	ErrInvitationNotAccepted     = errors.New("invitation is not accepted")
	ErrSetupTokenInvalid         = errors.New("setup token is invalid")
	ErrSetupTokenExpired         = errors.New("setup token has expired")
	ErrPasswordMismatch          = errors.New("passwords do not match")
	ErrEmailAlreadyExists        = errors.New("email already exists")
	ErrForbidden                 = errors.New("forbidden")
)

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
