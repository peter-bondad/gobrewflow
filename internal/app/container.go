package app

import (
	"gobrewflow/internal/config"
	"gobrewflow/internal/services/account"
	"gobrewflow/internal/services/auth"
	"gobrewflow/internal/services/categories"
	"gobrewflow/internal/services/invitations"
	"gobrewflow/internal/services/user"

	"github.com/uptrace/bun"
)

type Container struct {
	InvitationHandler invitations.InvitationHandler

	InvitationService invitations.InvitationService

	InvitationRepo invitations.InvitationRepository

	UserRepo user.UserRepository

	UserHandler user.UserHandler

	AccountRepo account.AccountRepository

	JwtService *auth.JWTService

	TokenBlacklistRepo auth.TokenBlacklistRepository

	CategoriesHandler categories.CategoryHandler
}

func NewContainer(db *bun.DB, cfg *config.Config) *Container {

	jwtService := &auth.JWTService{
		Secret: []byte(cfg.JWT.Secret),
	}

	tokenBlacklistRepo := auth.NewTokenBlacklistRepository(db)

	userRepo := user.NewUserRepository(db)
	userService := user.NewUserService(userRepo, jwtService, tokenBlacklistRepo)
	userHandler := user.NewUserHandler(userService)
	accountRepo := account.NewAccountRepository(db)
	invitationRepo := invitations.NewInvitationRepository(db)
	invitationService := invitations.NewInvitationService(
		db,
		invitationRepo,
		userRepo,
		accountRepo,
		cfg.Invitation,
	)

	invitationHandler := invitations.NewInvitationHandler(
		invitationService,
	)

	categoriesRepo := categories.NewCategoryRepository(db)
	categoriesService := categories.NewCategoryService(categoriesRepo)
	categoriesHandler := categories.NewCategoryHandler(categoriesService)

	return &Container{
		InvitationRepo:    invitationRepo,
		InvitationService: invitationService,
		InvitationHandler: invitationHandler,

		UserRepo:    userRepo,
		UserHandler: userHandler,

		AccountRepo: accountRepo,

		JwtService: jwtService,

		TokenBlacklistRepo: tokenBlacklistRepo,
		CategoriesHandler:  categoriesHandler,
	}
}
