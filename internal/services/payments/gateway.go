package payments

import (
	"context"

	"github.com/google/uuid"
)

type CreateCheckoutInput struct {
	ReferenceID string
	Amount      int64
	Currency    string
	Description string

	SuccessURL string
	CancelURL  string
}

type CreateCheckoutOutput struct {
	PaymentID   uuid.UUID
	CheckoutID  string
	CheckoutURL string
	Status      Status
}

type PaymentGateway interface {
	CreateCheckout(
		ctx context.Context,
		input CreateCheckoutInput,
	) (*CreateCheckoutOutput, error)
}
