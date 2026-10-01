package payments

import (
	"context"
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
