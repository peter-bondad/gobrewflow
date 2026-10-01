package payments

import "errors"

var (
	ErrInvalidOrderID  = errors.New("invalid order ID")
	ErrInvalidAmount   = errors.New("invalid payment amount")
	ErrInvalidCurrency = errors.New("invalid payment currency")
)
