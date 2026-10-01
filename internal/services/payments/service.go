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

	HandleWebhook(
		ctx context.Context,
		result *WebhookResult,
	) error
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

	// 2. Create the payment checkout through the payment gateway.
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

	// 4. Persist the payment.
	if err := s.repo.Create(ctx, s.db, payment); err != nil {
		return nil, err
	}

	// 5. Return the checkout information to the API/frontend.
	return &CreateCheckoutOutput{
		PaymentID:   payment.ID,
		CheckoutID:  checkout.CheckoutID,
		CheckoutURL: checkout.CheckoutURL,
		Status:      payment.Status,
	}, nil
}

// HandleWebhook processes a verified payment webhook.
// HandleWebhook processes a verified payment webhook.
func (s *paymentService) HandleWebhook(
	ctx context.Context,
	result *WebhookResult,
) error {
	// 1. Make sure we received a valid webhook result.
	if result == nil {
		return ErrInvalidWebhook
	}

	// 2. Find our payment using the PayMongo checkout ID.
	payment, err := s.repo.FindByProviderCheckoutID(
		ctx,
		s.db,
		result.CheckoutID,
	)
	if err != nil {
		return err
	}

	// 3. Ignore the webhook if this payment was already completed.
	// PayMongo can send the same webhook more than once.
	if payment.Status == StatusPaid {
		return nil
	}

	// 4. Only process successful payment events.
	if result.Status != StatusPaid {
		return nil
	}

	// 5. Update the payment with PayMongo's payment information.
	payment.Status = StatusPaid
	payment.ProviderPaymentID = result.PaymentID
	payment.UpdatedAt = time.Now()

	// 6. Persist the updated payment.
	return s.repo.Update(ctx, s.db, payment)
}
