package payments

import "errors"

var (
	ErrInvalidOrderID          = errors.New("invalid order ID")
	ErrInvalidAmount           = errors.New("invalid payment amount")
	ErrInvalidCurrency         = errors.New("invalid payment currency")
	ErrInvalidWebhook          = errors.New("invalid webhook payload")
	ErrInvalidWebhookSignature = errors.New("invalid webhook signature")
)
