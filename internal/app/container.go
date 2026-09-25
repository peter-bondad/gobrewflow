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

	// User
	userRepo := user.NewUserRepository(db)
	userService := user.NewUserService(
		userRepo,
		jwtService,
		tokenBlacklistRepo,
	)
	userHandler := user.NewUserHandler(userService)

	// Account
	accountRepo := account.NewAccountRepository(db)

	// Invitations
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

	// Categories
	categoriesRepo := categories.NewCategoryRepository(db)
	categoriesService := categories.NewCategoryService(categoriesRepo)
	categoriesHandler := categories.NewCategoryHandler(categoriesService)

	// Inventory
	inventoryRepo := inventory.NewInventoryRepository(db)

	// Inventory Movements
	inventoryMovementsRepo := inventory_movements.NewInventoryMovementsRepository(db)
	inventoryMovementService := inventory_movements.NewInventoryMovementsService(
		inventoryMovementsRepo,
	)

	// Inventory Service
	inventoryService := inventory.NewInventoryService(
		inventoryRepo,
		inventoryMovementService,
		txManager,
	)
	inventoryHandler := inventory.NewInventoryHandler(inventoryService)

	// Products
	productsRepo := products.NewProductRepository(db)
	productService := products.NewProductService(
		productsRepo,
		categoriesRepo,
		inventoryService,
		txManager,
	)
	productsHandler := products.NewProductHandler(productService)

	// Order Items
	orderItemsRepo := order_items.NewOrderItemRepository(db)
	orderItemsService := order_items.NewOrderItemsService(orderItemsRepo)

	// Orders
	ordersRepo := orders.NewOrderItemRepository(db)
	ordersService := orders.NewOrderService(
		ordersRepo,
		productsRepo,
		inventoryRepo,
		inventoryMovementsRepo,
		orderItemsService,
	)
	ordersHandler := orders.NewOrdesHandler(ordersService)

	return &Container{
		InvitationHandler:         invitationHandler,
		InvitationService:         invitationService,
		InvitationRepo:            invitationRepo,
		UserRepo:                  userRepo,
		UserHandler:               userHandler,
		AccountRepo:               accountRepo,
		JwtService:                jwtService,
		TokenBlacklistRepo:        tokenBlacklistRepo,
		TXManager:                 txManager,
		CategoriesHandler:         categoriesHandler,
		ProductsHandler:           productsHandler,
		InventoryHandler:          inventoryHandler,
		InventoryMovementsService: inventoryMovementService,
		OrdersHandler:             ordersHandler,
		OrdersService:             ordersService,
		OrdersRepo:                ordersRepo,
		OrderItemsService:         orderItemsService,
		OrderItemsRepo:            orderItemsRepo,
	}
}
