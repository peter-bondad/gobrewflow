package products

import "errors"

var (
	ProductIdIsRequired    = errors.New("product id is required")
	ProductInputIsRequired = errors.New("product input is required")
	ProductNameIsRequired  = errors.New("product name is required")
	ProductSKUIsRequired   = errors.New("product sku is required")
	ErrProductNotFound     = errors.New("product not found")
)
