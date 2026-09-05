package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gobrewflow/internal/services/auth"
)

type UserService interface {
	Login(ctx context.Context, input LoginInput) (string, error)
	Logout(ctx context.Context, tokenString string) error
	ListUsers(ctx context.Context, input UserListParams) ([]UserListItem, error)
}

type userService struct {
	repo               UserRepository
	auth               *auth.JWTService
	tokenBlacklistRepo auth.TokenBlacklistRepository
}

// NewUserService is a constructor function for the userService struct. It takes a UserRepository interface as an argument and returns a UserService interface. This allows for dependency injection and makes it easier to mock the service in tests.
func NewUserService(repo UserRepository, auth *auth.JWTService, tokenBlacklistRepo auth.TokenBlacklistRepository) UserService {
	return &userService{
		repo:               repo,
		auth:               auth,
		tokenBlacklistRepo: tokenBlacklistRepo,
	}
}

func (s userService) Login(
	ctx context.Context,
	input LoginInput,
) (string, error) {
	user, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return "", InvalidCredentials
	}

	token, _, err := s.auth.GenerateToken(user.ID.String())
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}

func (s userService) Logout(
	ctx context.Context,
	tokenString string,
) error {
	claims, err := s.auth.ParseToken(tokenString)
	if err != nil {
		return err
	}

	jti, ok := s.auth.ExtractJTI(claims)
	if !ok {
		return fmt.Errorf("missing jti in token")
	}

	expUnix, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("missing exp in token")
	}
	expiresAt := time.Unix(int64(expUnix), 0)

	return s.tokenBlacklistRepo.RevokeToken(ctx, jti, expiresAt)
}

type LoginInput struct {
	Email    string
	Password string
}

func (s *userService) ListUsers(
	ctx context.Context,
	params UserListParams,
) ([]UserListItem, error) {

	params.FullName = strings.TrimSpace(params.FullName)

	// Default pagination
	if params.Limit <= 0 {
		params.Limit = 10
	}

	// Maximum page size
	if params.Limit > 100 {
		params.Limit = 100
	}

	// Offset cannot be negative
	if params.Offset < 0 {
		params.Offset = 0
	}

	// Validate role
	if params.UserRole != nil {
		switch *params.UserRole {
		case Owner, Manager, Staff:
			// valid
		default:
			return nil, ErrInvalidUserRole
		}
	}

	params = UserListParams{
		FullName: params.FullName,
		Limit:    params.Limit,
		Offset:   params.Offset,
		UserRole: params.UserRole,
	}

	return s.repo.ListUsers(ctx, params)

}
