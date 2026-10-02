package payments

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type fakeOrdersService struct {
	markedOrderID uuid.UUID
	markErr       error
}

func (f *fakeOrdersService) MarkOrderAsPaid(
	ctx context.Context,
	orderID uuid.UUID,
) error {
	if f.markErr != nil {
		return f.markErr
	}

	f.markedOrderID = orderID
	return nil
}

type fakePaymentRepository struct {
	createdPayment *Payment
	createErr      error
	updatedPayment *Payment
	payment        *Payment
}

func (f *fakePaymentRepository) Create(
	ctx context.Context,
	db bun.IDB,
	payment *Payment,
) error {
	if f.createErr != nil {
		return f.createErr
	}

	f.createdPayment = payment
	return nil
}

func (f *fakePaymentRepository) FindByID(
	ctx context.Context,
	db bun.IDB,
	id uuid.UUID,
) (*Payment, error) {
	return nil, nil
}

func (f *fakePaymentRepository) FindByOrderID(
	ctx context.Context,
	db bun.IDB,
	orderID uuid.UUID,
) (*Payment, error) {
	return nil, nil
}
func (f *fakePaymentRepository) FindByProviderCheckoutID(
	ctx context.Context,
	db bun.IDB,
	checkoutID string,
) (*Payment, error) {
	return f.payment, nil
}

func (f *fakePaymentRepository) Update(
	ctx context.Context,
	db bun.IDB,
	payment *Payment,
) error {
	f.updatedPayment = payment

	return nil
}

type fakePaymentGateway struct {
	checkout  *CreateCheckoutOutput
	createErr error
}

func (f *fakePaymentGateway) CreateCheckout(
	ctx context.Context,
	input CreateCheckoutInput,
) (*CreateCheckoutOutput, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}

	return f.checkout, nil
}

const (
	testSuccessURL = "http://localhost:3000/payment/success"
	testCancelURL  = "http://localhost:3000/payment/cancel"
)

func TestPaymentService_CreateCheckout(t *testing.T) {
	orderID := uuid.New()

	t.Run("invalid order ID", func(t *testing.T) {
		service := NewService(
			&fakePaymentRepository{},
			&fakeOrdersService{},
			&fakePaymentGateway{},
			nil,
			testSuccessURL,
			testCancelURL,
		)

		_, err := service.CreateCheckout(
			context.Background(),
			CreatePaymentCheckoutInput{
				OrderID:  uuid.Nil,
				Amount:   10000,
				Currency: "PHP",
			},
		)

		if !errors.Is(err, ErrInvalidOrderID) {
			t.Fatalf("expected ErrInvalidOrderID, got %v", err)
		}
	})

	t.Run("invalid amount", func(t *testing.T) {
		service := NewService(
			&fakePaymentRepository{},
			&fakeOrdersService{},
			&fakePaymentGateway{},
			nil,
			testSuccessURL,
			testCancelURL,
		)

		_, err := service.CreateCheckout(
			context.Background(),
			CreatePaymentCheckoutInput{
				OrderID:  orderID,
				Amount:   0,
				Currency: "PHP",
			},
		)

		if !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("invalid currency", func(t *testing.T) {
		service := NewService(
			&fakePaymentRepository{},
			&fakeOrdersService{},
			&fakePaymentGateway{},
			nil,
			testSuccessURL,
			testCancelURL,
		)

		_, err := service.CreateCheckout(
			context.Background(),
			CreatePaymentCheckoutInput{
				OrderID:  orderID,
				Amount:   10000,
				Currency: "",
			},
		)

		if !errors.Is(err, ErrInvalidCurrency) {
			t.Fatalf("expected ErrInvalidCurrency, got %v", err)
		}
	})

	t.Run("gateway error", func(t *testing.T) {
		gatewayErr := errors.New("gateway error")

		service := NewService(
			&fakePaymentRepository{},
			&fakeOrdersService{},
			&fakePaymentGateway{
				createErr: gatewayErr,
			},
			nil,
			testSuccessURL,
			testCancelURL,
		)

		_, err := service.CreateCheckout(
			context.Background(),
			CreatePaymentCheckoutInput{
				OrderID:  orderID,
				Amount:   10000,
				Currency: "PHP",
			},
		)

		if !errors.Is(err, gatewayErr) {
			t.Fatalf("expected gateway error, got %v", err)
		}
	})

	t.Run("successful checkout", func(t *testing.T) {
		repo := &fakePaymentRepository{}

		gateway := &fakePaymentGateway{
			checkout: &CreateCheckoutOutput{
				CheckoutID:  "cs_test_123",
				CheckoutURL: "https://checkout.paymongo.com/test",
			},
		}

		service := NewService(repo, &fakeOrdersService{}, gateway, nil, testSuccessURL, testCancelURL)

		result, err := service.CreateCheckout(
			context.Background(),
			CreatePaymentCheckoutInput{
				OrderID:     orderID,
				ReferenceID: "ORDER-001",
				Amount:      10000,
				Currency:    "PHP",
				Description: "Test Order",
			},
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("expected result, got nil")
		}

		if result.PaymentID == uuid.Nil {
			t.Fatal("expected payment ID")
		}

		if result.CheckoutID != "cs_test_123" {
			t.Fatalf("expected checkout ID cs_test_123, got %s", result.CheckoutID)
		}

		if result.CheckoutURL != "https://checkout.paymongo.com/test" {
			t.Fatalf("unexpected checkout URL: %s", result.CheckoutURL)
		}

		if result.Status != StatusPending {
			t.Fatalf("expected status %s, got %s", StatusPending, result.Status)
		}

		if repo.createdPayment == nil {
			t.Fatal("expected payment to be persisted")
		}

		if repo.createdPayment.OrderID != orderID {
			t.Fatalf("expected order ID %s, got %s",
				orderID,
				repo.createdPayment.OrderID,
			)
		}

		if repo.createdPayment.Provider != ProviderPayMongo {
			t.Fatalf(
				"expected provider %s, got %s",
				ProviderPayMongo,
				repo.createdPayment.Provider,
			)
		}

		if repo.createdPayment.ProviderCheckoutID != "cs_test_123" {
			t.Fatalf(
				"expected checkout ID cs_test_123, got %s",
				repo.createdPayment.ProviderCheckoutID,
			)
		}

		if repo.createdPayment.Amount != 10000 {
			t.Fatalf(
				"expected amount 10000, got %d",
				repo.createdPayment.Amount,
			)
		}

		if repo.createdPayment.Currency != "PHP" {
			t.Fatalf(
				"expected currency PHP, got %s",
				repo.createdPayment.Currency,
			)
		}

		if repo.createdPayment.Status != StatusPending {
			t.Fatalf(
				"expected payment status %s, got %s",
				StatusPending,
				repo.createdPayment.Status,
			)
		}
	})
}

func TestPaymentService_HandleWebhook(t *testing.T) {

	checkoutID := "cs_test_123"
	providerPaymentID := "pay_test_123"

	t.Run("successful payment updates payment", func(t *testing.T) {
		payment := &Payment{
			ID:                 uuid.New(),
			ProviderCheckoutID: checkoutID,
			Status:             StatusPending,
		}

		repo := &fakePaymentRepository{
			payment: payment,
		}

		service := NewService(
			repo,
			&fakeOrdersService{},
			nil,
			nil,
			"",
			"",
		)

		err := service.HandleWebhook(
			context.Background(),
			&WebhookResult{
				CheckoutID: checkoutID,
				PaymentID:  &providerPaymentID,
				Status:     StatusPaid,
			},
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if repo.updatedPayment == nil {
			t.Fatal("expected payment to be updated")
		}

		if repo.updatedPayment.Status != StatusPaid {
			t.Fatalf(
				"expected status %s, got %s",
				StatusPaid,
				repo.updatedPayment.Status,
			)
		}

		if repo.updatedPayment.ProviderPaymentID == nil {
			t.Fatal("expected provider payment ID")
		}

		if *repo.updatedPayment.ProviderPaymentID != providerPaymentID {
			t.Fatalf(
				"expected provider payment ID %s, got %s",
				providerPaymentID,
				*repo.updatedPayment.ProviderPaymentID,
			)
		}
	})

	t.Run("already paid payment is ignored", func(t *testing.T) {
		payment := &Payment{
			ID:                 uuid.New(),
			ProviderCheckoutID: checkoutID,
			Status:             StatusPaid,
		}

		repo := &fakePaymentRepository{
			payment: payment,
		}

		service := NewService(
			repo,
			&fakeOrdersService{},
			nil,
			nil,
			"",
			"",
		)

		err := service.HandleWebhook(
			context.Background(),
			&WebhookResult{
				CheckoutID: checkoutID,
				PaymentID:  &providerPaymentID,
				Status:     StatusPaid,
			},
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if repo.updatedPayment != nil {
			t.Fatal("expected already paid payment not to be updated")
		}
	})

	t.Run("non-paid webhook is ignored", func(t *testing.T) {
		payment := &Payment{
			ID:                 uuid.New(),
			ProviderCheckoutID: checkoutID,
			Status:             StatusPending,
		}

		repo := &fakePaymentRepository{
			payment: payment,
		}

		service := NewService(
			repo,
			&fakeOrdersService{},
			nil,
			nil,
			"",
			"",
		)

		err := service.HandleWebhook(
			context.Background(),
			&WebhookResult{
				CheckoutID: checkoutID,
				PaymentID:  &providerPaymentID,
				Status:     StatusFailed,
			},
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if repo.updatedPayment != nil {
			t.Fatal("expected non-paid webhook not to update payment")
		}
	})
}
