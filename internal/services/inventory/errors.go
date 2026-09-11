package inventory

import "errors"

var (
	ErrInventoryNotFound = errors.New("inventory not found")
	ErrInvalidQuantity   = errors.New("quantity must be greater than zero")
	ErrInsufficientStock = errors.New("insufficient stock")
)
