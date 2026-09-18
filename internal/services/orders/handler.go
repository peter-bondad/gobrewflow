package orders

import (
	"gobrewflow/internal/services/auth"
	"gobrewflow/internal/services/order_items"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrdersHandler interface {
	CreateOrder(c *gin.Context)
}

type ordersHandler struct {
	service OrdersService
}

func NewOrdesHandler(service OrdersService) OrdersHandler {
	return &ordersHandler{
		service: service,
	}
}

type CreateOrderRequest struct {
	Items []CreateOrderItemRequest `json:"items" binding:"required,min=1"`
}

type CreateOrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
}

func (h *ordersHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// Get authenticated user from middleware.
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	items := make([]order_items.OrderProductItem, len(req.Items))

	for i, item := range req.Items {
		items[i] = order_items.OrderProductItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	input := &CreateOrderInput{
		CashierID: userID,
		Items:     items,
	}

	order, err := h.service.CreateOrder(c, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, order)
}
