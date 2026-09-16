package orders

import "errors"

var (
	ErrInvalidOrderID   = errors.New("invalid order id")
	ErrInvalidCashierID = errors.New("invalid cashier id")
)
