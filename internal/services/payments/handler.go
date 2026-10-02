package payments

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"gobrewflow/internal/services/orders"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentHandler interface {
	CreateCheckout(c *gin.Context)
	HandlePayMongoWebhook(c *gin.Context)
}

type OrderFinder interface {
	FindByID(ctx context.Context, id uuid.UUID) (*orders.OrderOutput, error)
}

type paymentHandler struct {
	log             *slog.Logger
	paymentService  PaymentService
	orderService    OrderFinder
	webhookVerifier PayMongoWebhookVerifier
}

func NewPaymentHandler(
	paymentService PaymentService,
	orderService OrderFinder,
	webhookVerifier PayMongoWebhookVerifier,
	log *slog.Logger,
) PaymentHandler {
	return &paymentHandler{
		paymentService:  paymentService,
		orderService:    orderService,
		webhookVerifier: webhookVerifier,
		log:             log,
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

// CreateCheckout creates a PayMongo checkout session for an order.
func (h *paymentHandler) CreateCheckout(c *gin.Context) {
	var req CreateCheckoutRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// 1. Find the order that the customer wants to pay for.
	order, err := h.orderService.FindByID(c, req.OrderID)
	if err != nil {
		c.Error(err)
		return
	}

	// 2. Make sure the order is still waiting for payment.
	if order.Status != string(orders.OrderStatusPending) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "order is not payable",
		})
		return
	}

	// 3. Ask the payment service to create the checkout.
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

	// 4. Return the checkout information to the client.
	resp := CreateCheckoutResponse{
		PaymentID:   payment.PaymentID,
		CheckoutID:  payment.CheckoutID,
		CheckoutURL: payment.CheckoutURL,
		Status:      string(payment.Status),
	}

	c.JSON(http.StatusCreated, resp)
}

type WebhookResult struct {
	CheckoutID      string
	PaymentID       *string
	ReferenceNumber string
	Amount          int64
	Currency        string
	Status          Status
	PaymentMethod   *PaymentMethod
}

type PayMongoWebhookVerifier interface {
	VerifyWebhookSignature(payload []byte, signature string) error
	ParseWebhook(payload []byte) (*WebhookResult, error)
}

// HandlePayMongoWebhook receives payment events sent by PayMongo.
func (h *paymentHandler) HandlePayMongoWebhook(c *gin.Context) {
	// 1. Read the raw request body.
	// We need the exact body to verify PayMongo's signature.
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to read webhook body",
		})
		return
	}

	h.log.Info(
		"paymongo_webhook_received",
		"payload", string(payload),
	)

	// 2. Get the signature sent by PayMongo.
	signature := c.GetHeader("Paymongo-Signature")

	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "missing PayMongo signature",
		})
		return
	}

	// 3. Verify that the webhook was signed by PayMongo.
	if err := h.webhookVerifier.VerifyWebhookSignature(
		payload,
		signature,
	); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid webhook signature",
		})
		return
	}

	// 4. Parse the verified webhook payload.
	result, err := h.webhookVerifier.ParseWebhook(payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 5. Pass the payment information to the payment service.
	// The service will handle the database/business logic.
	if err := h.paymentService.HandleWebhook(c, result); err != nil {
		c.Error(err)
		return
	}

	// 6. Tell PayMongo that BrewFlow successfully processed the webhook.
	c.Status(http.StatusOK)
}
