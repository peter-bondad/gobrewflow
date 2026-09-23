package server

import (
	"gobrewflow/internal/middleware"
	"gobrewflow/internal/services/categories"
	"gobrewflow/internal/services/inventory"
	"gobrewflow/internal/services/invitations"
	"gobrewflow/internal/services/orders"
	"gobrewflow/internal/services/products"
	"gobrewflow/internal/services/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func (s *Server) routes() {
	s.Server.GET("/health", s.handleHealth)
}

func (s *Server) publicRoutes(userHandler user.UserHandler, invitationHandler invitations.InvitationHandler) {

	api := s.Server.Group("/api")
	api.POST("/login", userHandler.Login)
	api.POST("/logout", userHandler.Logout)

	// invitation for user routes
	api.POST("/accept-invitation", invitationHandler.AcceptInvitation)
	api.POST("/set-password", invitationHandler.SetPassword)
}

// internal/server/routes.go
func (s *Server) protectedRoutes(
	authMiddleware *middleware.AuthMiddleware,
	userRepo user.UserRepository,
	userHandler user.UserHandler,
	invitationHandler invitations.InvitationHandler,
	categoriesHandler categories.CategoryHandler,
	productsHandler products.ProductHandler,
	inventoryHandler inventory.InventoryHandler,
	ordersHandler orders.OrdersHandler,
) {
	protectedAPI := s.Server.Group("/api/protected")

	protectedAPI.Use(authMiddleware.AuthHandler())

	userAPI := protectedAPI.Group("/users")
	invitationAPI := protectedAPI.Group("/invitations")
	categoriesAPI := protectedAPI.Group("/categories")
	productsAPI := protectedAPI.Group("/products")
	inventoryAPI := protectedAPI.Group("/inventory")
	ordersAPI := protectedAPI.Group("/orders")
	userAPI.Use(
		middleware.RequireRoles(
			userRepo,
			user.Owner,
			user.Manager,
		),
	)
	invitationAPI.Use(
		middleware.RequireRoles(
			userRepo,
			user.Owner,
			user.Manager,
		),
	)
	// Users
	userAPI.GET("/", userHandler.ListUsers)

	// Invitations
	invitationAPI.POST("/", invitationHandler.SendInvitation)
	invitationAPI.POST("/:id/cancel", invitationHandler.CancelInvitation)
	invitationAPI.GET("/", invitationHandler.ListInvitations)

	// Categories
	categoriesAPI.POST("/", categoriesHandler.CreateCategory)
	categoriesAPI.PATCH("/:id", categoriesHandler.UpdateCategoryName)
	categoriesAPI.PATCH("/:id/status", categoriesHandler.SetCategoryStatus)
	categoriesAPI.GET("/", categoriesHandler.ListCategories)

	// Products
	productsAPI.POST("/", productsHandler.CreateProducts)
	productsAPI.GET("/", productsHandler.ListProducts)
	productsAPI.GET("/sku/:sku", productsHandler.FindProductBySKU)
	productsAPI.GET("/:id", productsHandler.FindProductByID)

	// Inventory
	inventoryAPI.GET("/:productId", inventoryHandler.GetInventoryByProductID)
	inventoryAPI.PATCH("/:productId/adjust", inventoryHandler.AdjustStock)

	// Orders
	ordersAPI.POST("/", ordersHandler.CreateOrder)
}
