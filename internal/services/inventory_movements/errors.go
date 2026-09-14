package inventory_movements

import "errors"

var (
	ErrInvalidMovementType = errors.New("invalid movement type")
	ErrInvalidQuantity     = errors.New("quantity must be greater than zero")
	ErrProductNotFound     = errors.New("product not found")
	ErrInventoryNotFound   = errors.New("inventory not found")
	ErrInsufficientStock   = errors.New("insufficient stock")
	ErrMovementNotFound    = errors.New("movement not found")
	InvalidLimitParameter  = errors.New("invalid limit parameter")
	InvalidPageParameter   = errors.New("invalid page parameter")
)
