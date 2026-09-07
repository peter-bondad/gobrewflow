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
	ListUsers(ctx context.Context, input ListUserInput) (ListUserOutput, error)
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

type LoginInput struct {
	Email    string
	Password string
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

type ListUserInput struct {
	FullName string
	Email    string
	Role     *UserRole
	Page     int
	Limit    int
}

type ListUserOutput struct {
	Data       []UserListItem
	Page       int
	Limit      int
	Total      int
	TotalPages int
}

func (s *userService) ListUsers(
	ctx context.Context,
	input ListUserInput,
) (ListUserOutput, error) {

	input.FullName = strings.TrimSpace(input.FullName)
	input.Email = strings.TrimSpace(input.Email)

	// Default pagination
	if input.Limit <= 0 {
		input.Limit = 10
	}

	// Maximum page size
	if input.Limit > 100 {
		input.Limit = 100
	}

	// Default page
	if input.Page <= 0 {
		input.Page = 1
	}

	// Validate role
	if input.Role != nil {
		switch *input.Role {
		case Owner, Manager, Staff:
			// valid
		default:
			return ListUserOutput{}, ErrInvalidUserRole
		}
	}

	offset := (input.Page - 1) * input.Limit

	params := UserListParams{
		FullName: input.FullName,
		Email:    input.Email,
		UserRole: input.Role,
		Limit:    input.Limit,
		Offset:   offset,
	}

	result, err := s.repo.ListUsers(ctx, params)
	if err != nil {
		return ListUserOutput{}, err
	}

	totalPages := 0
	if result.Total > 0 {
		totalPages = (result.Total + input.Limit - 1) / input.Limit
	}

	return ListUserOutput{
		Data:       result.Data,
		Page:       input.Page,
		Limit:      input.Limit,
		Total:      result.Total,
		TotalPages: totalPages,
	}, nil
}
