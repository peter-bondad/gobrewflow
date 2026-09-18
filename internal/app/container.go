package app

import (
	"gobrewflow/internal/config"
	"gobrewflow/internal/database"
	"gobrewflow/internal/services/account"
	"gobrewflow/internal/services/auth"
	"gobrewflow/internal/services/categories"
	"gobrewflow/internal/services/inventory"
	"gobrewflow/internal/services/inventory_movements"
	"gobrewflow/internal/services/invitations"
	"gobrewflow/internal/services/order_items"
	"gobrewflow/internal/services/orders"
	"gobrewflow/internal/services/products"
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

	TXManager database.TxManager

	CategoriesHandler categories.CategoryHandler

	ProductsHandler products.ProductHandler

	InventoryHandler inventory.InventoryHandler

	InventoryMovementsService inventory_movements.InventoryMovementsService

	OrdersHandler orders.OrdersHandler

	OrdersService orders.OrdersService

	OrdersRepo orders.OrderRepository

	OrderItemsService order_items.OrderItemsService

	OrderItemsRepo order_items.OrderItemRepository
}

func NewContainer(db *bun.DB, cfg *config.Config) *Container {

	jwtService := &auth.JWTService{
		Secret: []byte(cfg.JWT.Secret),
	}

	tokenBlacklistRepo := auth.NewTokenBlacklistRepository(db)

	txManager := database.NewTxManager(db)

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

	inventoryRepo := inventory.NewInventoryRepository(db)
	inventoryService := inventory.NewInventoryService(inventoryRepo)
	inventoryHandler := inventory.NewInventoryHandler(inventoryService)

	productsRepo := products.NewProductRepository(db)
	productService := products.NewProductService(productsRepo, categoriesRepo, inventoryService, txManager)
	productsHandler := products.NewProductHandler(productService)

	inventoryMovementsRepo := inventory_movements.NewInventoryMovementsRepository(db)
	inventoryMovementService := inventory_movements.NewInventoryMovementsService(
		inventoryMovementsRepo,
		inventoryRepo,
	)

	orderItemsRepo := order_items.NewOrderItemRepository(db)
	orderItemsService := order_items.NewOrderItemsService(orderItemsRepo)

	ordersRepo := orders.NewOrderItemRepository(db)
	ordersService := orders.NewOrderService(ordersRepo, productsRepo, inventoryRepo, inventoryMovementsRepo, orderItemsService)
	ordersHandler := orders.NewOrdesHandler(ordersService)

	return &Container{
		InvitationRepo:    invitationRepo,
		InvitationService: invitationService,
		InvitationHandler: invitationHandler,

		UserRepo:    userRepo,
		UserHandler: userHandler,

		AccountRepo: accountRepo,

		JwtService: jwtService,

		TXManager: txManager,

		TokenBlacklistRepo: tokenBlacklistRepo,
		CategoriesHandler:  categoriesHandler,
		ProductsHandler:    productsHandler,

		InventoryHandler:          inventoryHandler,
		InventoryMovementsService: inventoryMovementService,

		OrdersHandler: ordersHandler,
		OrdersService: ordersService,
		OrdersRepo:    ordersRepo,

		OrderItemsService: orderItemsService,
		OrderItemsRepo:    orderItemsRepo,
	}
}
