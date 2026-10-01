package payments

import (
	"context"
	"net/http"

	"gobrewflow/internal/services/orders"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentHandler interface {
	CreateCheckout(c *gin.Context)
}

type OrderFinder interface {
	FindByID(ctx context.Context, id uuid.UUID) (*orders.OrderOutput, error)
}

type paymentHandler struct {
	paymentService PaymentService
	orderService   OrderFinder
}

func NewPaymentHandler(
	paymentService PaymentService,
	orderService OrderFinder,
) PaymentHandler {
	return &paymentHandler{
		paymentService: paymentService,
		orderService:   orderService,
	}
}

type CreateCheckoutRequest struct {
	OrderID uuid.UUID `json:"order_id" binding:"required"`
}

type CreateCheckoutResponse struct {
	PaymentID   uuid.UUID `json:"payment_id"`
	CheckoutID  string    `json:"checkout_id"`
	CheckoutURL string    `json:"checkout_url"`
	Status      string    `json:"status"`
}

func (h *paymentHandler) CreateCheckout(c *gin.Context) {
	var req CreateCheckoutRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	order, err := h.orderService.FindByID(c, req.OrderID)
	if err != nil {
		c.Error(err)
		return
	}

	if order.Status != string(orders.OrderStatusPending) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "order is not payable",
		})
		return
	}

	payment, err := h.paymentService.CreateCheckout(
		c,
		CreatePaymentCheckoutInput{
			OrderID:     order.ID,
			ReferenceID: order.OrderNumber,
			Amount:      order.Total,
			Currency:    "PHP",
			Description: "Order " + order.OrderNumber,
		},
	)
	if err != nil {
		c.Error(err)
		return
	}

	resp := CreateCheckoutResponse{
		PaymentID:   payment.PaymentID,
		CheckoutID:  payment.CheckoutID,
		CheckoutURL: payment.CheckoutURL,
		Status:      string(payment.Status),
	}

	c.JSON(http.StatusCreated, resp)
}
