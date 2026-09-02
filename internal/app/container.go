package app

import (
	"gobrewflow/internal/config"
	"gobrewflow/internal/services/account"
	"gobrewflow/internal/services/auth"
	"gobrewflow/internal/services/invitation"
	"gobrewflow/internal/services/user"

	"github.com/uptrace/bun"
)

type Container struct {
	InvitationHandler invitation.InvitationHandler

	InvitationService invitation.InvitationService

	InvitationRepo invitation.InvitationRepository

	UserRepo user.UserRepository

	UserHandler user.UserHandler

	AccountRepo account.AccountRepository

	JwtService *auth.JWTService

	TokenBlacklistRepo auth.TokenBlacklistRepository
}

func NewContainer(db *bun.DB, cfg *config.Config) *Container {

	jwtService := &auth.JWTService{
		Secret: []byte(cfg.JWT.Secret),
	}

	userRepo := user.NewUserRepository(db)
	accountRepo := account.NewAccountRepository(db)
	invitationRepo := invitation.NewInvitationRepository(db)
	tokenBlacklistRepo := auth.NewTokenBlacklistRepository(db)

	userService := user.NewUserService(userRepo, jwtService, tokenBlacklistRepo)

	userHandler := user.NewUserHandler(userService)
	invitationService := invitation.NewInvitationService(
		db,
		invitationRepo,
		userRepo,
		accountRepo,
		cfg.Invitation,
	)

	invitationHandler := invitation.NewInvitationHandler(
		invitationService,
	)

	return &Container{
		InvitationRepo:    invitationRepo,
		InvitationService: invitationService,
		InvitationHandler: invitationHandler,

		UserRepo:    userRepo,
		UserHandler: userHandler,

		AccountRepo: accountRepo,

		JwtService: jwtService,

		TokenBlacklistRepo: tokenBlacklistRepo,
	}
}
