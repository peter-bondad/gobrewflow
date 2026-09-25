package inventory

import "errors"

var (
	ErrInventoryNotFound    = errors.New("inventory not found")
	ErrInvalidQuantity      = errors.New("quantity must be greater than zero")
	ErrInsufficientStock    = errors.New("insufficient stock")
	ErrInvalidStockChange   = errors.New("invalid stock change type")
	ErrStockOverflow        = errors.New("stock quantity exceeds maximum allowed value")
	ErrInvalidStockIncrease = errors.New("after stock must be greater than before stock")
)
