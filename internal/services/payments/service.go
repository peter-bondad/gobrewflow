package payments

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PaymentService interface {
	CreateCheckout(
		ctx context.Context,
		input CreatePaymentCheckoutInput,
	) (*CreateCheckoutOutput, error)
}

type paymentService struct {
	repo       Repository
	gateway    PaymentGateway
	db         bun.IDB
	successURL string
	cancelURL  string
}

func NewService(
	repo Repository,
	gateway PaymentGateway,
	db bun.IDB,
	successURL string,
	cancelURL string,
) PaymentService {
	return &paymentService{
		repo:       repo,
		gateway:    gateway,
		db:         db,
		successURL: successURL,
		cancelURL:  cancelURL,
	}
}

type CreatePaymentCheckoutInput struct {
	OrderID     uuid.UUID
	ReferenceID string
	Amount      int64
	Currency    string
	Description string
}

func (s *paymentService) CreateCheckout(
	ctx context.Context,
	input CreatePaymentCheckoutInput,
) (*CreateCheckoutOutput, error) {
	// 1. Validate input.
	if input.OrderID == uuid.Nil {
		return nil, ErrInvalidOrderID
	}

	if input.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if input.Currency == "" {
		return nil, ErrInvalidCurrency
	}

	// 2. Create the payment checkout.
	checkout, err := s.gateway.CreateCheckout(
		ctx,
		CreateCheckoutInput{
			ReferenceID: input.ReferenceID,
			Amount:      input.Amount,
			Currency:    input.Currency,
			Description: input.Description,
			SuccessURL:  s.successURL,
			CancelURL:   s.cancelURL,
		},
	)
	if err != nil {
		return nil, err
	}

	// 3. Create our internal payment record.
	now := time.Now()

	payment := &Payment{
		ID:                 uuid.New(),
		OrderID:            input.OrderID,
		Provider:           ProviderPayMongo,
		ProviderCheckoutID: checkout.CheckoutID,
		Amount:             input.Amount,
		Currency:           input.Currency,
		Status:             StatusPending,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	// 4. Persist it.
	if err := s.repo.Create(ctx, s.db, payment); err != nil {
		return nil, err
	}

	// 5. Return data needed by the API/frontend.
	return &CreateCheckoutOutput{
		PaymentID:   payment.ID,
		CheckoutID:  checkout.CheckoutID,
		CheckoutURL: checkout.CheckoutURL,
		Status:      payment.Status,
	}, nil
}
