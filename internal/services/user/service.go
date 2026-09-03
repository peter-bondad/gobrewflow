package user

import (
	"context"
	"fmt"
	"time"

	"gobrewflow/internal/services/auth"
)

type UserService interface {
	Login(ctx context.Context, input LoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context, tokenString string) error
	ListUsers(ctx context.Context, input UserListInput) ([]UserListItem, error)
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
	input LoginRequest,
) (*LoginResponse, error) {
	user, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, InvalidCredentials
	}

	token, _, err := s.auth.GenerateToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &LoginResponse{
		Token: token,
	}, nil
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

func (s *userService) ListUsers(
	ctx context.Context,
	input UserListInput,
) ([]UserListItem, error) {

	// Default pagination
	if input.Limit <= 0 {
		input.Limit = 10
	}

	// Maximum page size
	if input.Limit > 100 {
		input.Limit = 100
	}

	// Offset cannot be negative
	if input.Offset < 0 {
		input.Offset = 0
	}

	// Validate role
	if input.UserRole != nil {
		switch *input.UserRole {
		case Owner, Manager, Staff:
			// valid
		default:
			return nil, ErrInvalidUserRole
		}
	}

	return s.repo.ListUsers(ctx, input)
}
