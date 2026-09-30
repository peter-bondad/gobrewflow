package payments

import (
	"time"

	"github.com/google/uuid"
)

type Provider string

const (
	ProviderPayMongo Provider = "PAYMONGO"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusPaid      Status = "PAID"
	StatusFailed    Status = "FAILED"
	StatusCancelled Status = "CANCELLED"
	StatusRefunded  Status = "REFUNDED"
)

type PaymentMethod string

const (
	PaymentMethodCard    PaymentMethod = "CARD"
	PaymentMethodGCash   PaymentMethod = "GCASH"
	PaymentMethodGrabPay PaymentMethod = "GRABPAY"
	PaymentMethodMaya    PaymentMethod = "MAYA"
	PaymentMethodQRPH    PaymentMethod = "QRPH"
)

type Payment struct {
	ID      uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	OrderID uuid.UUID `bun:"order_id,type:uuid,notnull"`

	Provider Provider `bun:"provider,notnull"`

	ProviderCheckoutID string  `bun:"provider_checkout_id,notnull"`
	ProviderPaymentID  *string `bun:"provider_payment_id"`

	Amount   int64  `bun:"amount,notnull"`
	Currency string `bun:"currency,notnull"`

	Status        Status         `bun:"status,notnull"`
	PaymentMethod *PaymentMethod `bun:"payment_method"`

	PaidAt *time.Time `bun:"paid_at"`

	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull"`
}
