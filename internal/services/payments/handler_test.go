package payments

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gobrewflow/internal/services/orders"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeOrderService struct {
	order *orders.OrderOutput
	err   error
	log   *slog.Logger
}

func (f *fakeOrderService) CreateOrder(
	ctx context.Context,
	input *orders.CreateOrderInput,
) (*orders.CreateOrderOutput, error) {
	return nil, nil
}

func (f *fakeOrderService) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*orders.OrderOutput, error) {
	if f.err != nil {
		return nil, f.err
	}

	return f.order, nil
}

type fakePaymentService struct {
	result *CreateCheckoutOutput
	err    error

	input CreatePaymentCheckoutInput

	webhookResult *WebhookResult
	webhookErr    error
	log           *slog.Logger
}

func (f *fakePaymentService) CreateCheckout(
	ctx context.Context,
	input CreatePaymentCheckoutInput,
) (*CreateCheckoutOutput, error) {
	f.input = input

	if f.err != nil {
		return nil, f.err
	}

	return f.result, nil
}

func (f *fakePaymentService) HandleWebhook(
	ctx context.Context,
	result *WebhookResult,
) error {
	f.webhookResult = result

	return f.webhookErr
}

type fakeWebhookVerifier struct {
	verifyErr error
	parseErr  error
	result    *WebhookResult
}

func (f *fakeWebhookVerifier) VerifyWebhookSignature(
	payload []byte,
	signature string,
) error {
	return f.verifyErr
}

func (f *fakeWebhookVerifier) ParseWebhook(
	payload []byte,
) (*WebhookResult, error) {
	if f.parseErr != nil {
		return nil, f.parseErr
	}

	return f.result, nil
}

func stringPtr(value string) *string {
	return &value
}

func TestPaymentHandler_CreateCheckout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderID := uuid.New()

	orderService := &fakeOrderService{
		order: &orders.OrderOutput{
			ID:          orderID,
			OrderNumber: "ORD-001",
			Status:      string(orders.OrderStatusPending),
			Total:       10000,
		},
	}

	paymentService := &fakePaymentService{
		result: &CreateCheckoutOutput{
			PaymentID:   uuid.New(),
			CheckoutID:  "cs_test_123",
			CheckoutURL: "https://checkout.paymongo.com/test",
			Status:      StatusPending,
		},
	}

	handler := NewPaymentHandler(
		paymentService,
		orderService,
		&fakeWebhookVerifier{},
		slog.Default(),
	)

	router := gin.New()
	router.POST("/payments/checkout", handler.CreateCheckout)

	body := `{
		"order_id": "` + orderID.String() + `"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/payments/checkout",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}

	if !strings.Contains(rec.Body.String(), `"checkout_id":"cs_test_123"`) {
		t.Fatalf(
			"expected checkout ID in response, got %s",
			rec.Body.String(),
		)
	}

	if paymentService.input.OrderID != orderID {
		t.Fatalf(
			"expected order ID %s, got %s",
			orderID,
			paymentService.input.OrderID,
		)
	}

	if paymentService.input.Amount != 10000 {
		t.Fatalf(
			"expected amount 10000, got %d",
			paymentService.input.Amount,
		)
	}

	if paymentService.input.Currency != "PHP" {
		t.Fatalf(
			"expected currency PHP, got %s",
			paymentService.input.Currency,
		)
	}

	if paymentService.input.ReferenceID != "ORD-001" {
		t.Fatalf(
			"expected reference ID ORD-001, got %s",
			paymentService.input.ReferenceID,
		)
	}

	t.Run("completed order cannot create checkout", func(t *testing.T) {
		orderID := uuid.New()

		orderService := &fakeOrderService{
			order: &orders.OrderOutput{
				ID:          orderID,
				OrderNumber: "ORD-002",
				Status:      string(orders.OrderStatusCompleted),
				Total:       10000,
			},
		}

		paymentService := &fakePaymentService{}

		handler := NewPaymentHandler(
			paymentService,
			orderService,
			&fakeWebhookVerifier{},
			slog.Default(),
		)

		router := gin.New()
		router.POST("/payments/checkout", handler.CreateCheckout)

		body := `{
			"order_id": "` + orderID.String() + `"
		}`

		req := httptest.NewRequest(
			http.MethodPost,
			"/payments/checkout",
			strings.NewReader(body),
		)

		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				rec.Code,
			)
		}

		if paymentService.input.OrderID != uuid.Nil {
			t.Fatal("expected payment service not to be called")
		}
	})
}

func TestPaymentHandler_HandlePayMongoWebhook(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successfully handles webhook", func(t *testing.T) {
		paymentService := &fakePaymentService{}

		webhookVerifier := &fakeWebhookVerifier{
			result: &WebhookResult{
				CheckoutID: "cs_test_123",
				PaymentID:  stringPtr("pay_test_123"),
				Status:     StatusPaid,
			},
		}

		handler := NewPaymentHandler(
			paymentService,
			&fakeOrderService{},
			webhookVerifier,
			slog.Default(),
		)

		router := gin.New()
		router.POST(
			"/api/webhooks/paymongo",
			handler.HandlePayMongoWebhook,
		)

		body := `{"event_type":"checkout_session.payment.paid"}`

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/webhooks/paymongo",
			strings.NewReader(body),
		)

		req.Header.Set(
			"Paymongo-Signature",
			"t=123,te=test-signature",
		)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusOK,
				rec.Code,
			)
		}

		if paymentService.webhookResult == nil {
			t.Fatal("expected payment service to receive webhook result")
		}

		if paymentService.webhookResult.CheckoutID != "cs_test_123" {
			t.Fatalf(
				"expected checkout ID cs_test_123, got %s",
				paymentService.webhookResult.CheckoutID,
			)
		}

		if paymentService.webhookResult.PaymentID == nil {
			t.Fatal("expected payment ID")
		}

		if *paymentService.webhookResult.PaymentID != "pay_test_123" {
			t.Fatalf(
				"expected payment ID pay_test_123, got %s",
				*paymentService.webhookResult.PaymentID,
			)
		}

		if paymentService.webhookResult.Status != StatusPaid {
			t.Fatalf(
				"expected status %s, got %s",
				StatusPaid,
				paymentService.webhookResult.Status,
			)
		}
	})

	t.Run("missing signature returns bad request", func(t *testing.T) {
		paymentService := &fakePaymentService{}

		webhookVerifier := &fakeWebhookVerifier{
			result: &WebhookResult{
				CheckoutID: "cs_test_123",
				PaymentID:  stringPtr("pay_test_123"),
				Status:     StatusPaid,
			},
		}

		handler := NewPaymentHandler(
			paymentService,
			&fakeOrderService{},
			webhookVerifier,
			slog.Default(),
		)

		router := gin.New()
		router.POST(
			"/api/webhooks/paymongo",
			handler.HandlePayMongoWebhook,
		)

		body := `{"event_type":"checkout_session.payment.paid"}`

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/webhooks/paymongo",
			strings.NewReader(body),
		)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				rec.Code,
			)
		}

		if paymentService.webhookResult != nil {
			t.Fatal("expected payment service not to be called")
		}
	})

	t.Run("invalid signature returns unauthorized", func(t *testing.T) {
		paymentService := &fakePaymentService{}

		webhookVerifier := &fakeWebhookVerifier{
			verifyErr: ErrInvalidWebhookSignature,
		}

		handler := NewPaymentHandler(
			paymentService,

			&fakeOrderService{},
			webhookVerifier,
			slog.Default(),
		)

		router := gin.New()
		router.POST(
			"/api/webhooks/paymongo",
			handler.HandlePayMongoWebhook,
		)

		body := `{"event_type":"checkout_session.payment.paid"}`

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/webhooks/paymongo",
			strings.NewReader(body),
		)

		req.Header.Set(
			"Paymongo-Signature",
			"t=123,te=invalid-signature",
		)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusUnauthorized,
				rec.Code,
			)
		}

		if paymentService.webhookResult != nil {
			t.Fatal("expected payment service not to be called")
		}
	})
}
