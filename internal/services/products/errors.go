package products

import "errors"

var (
	ProductNameAlreadyExists       = errors.New("product name already exists")
	ProductIdIsRequired            = errors.New("product id is required")
	ProductInputIsRequired         = errors.New("product input is required")
	ProductNameCannotBeEmpty       = errors.New("product name cannot be empty")
	ProductNameTooLong             = errors.New("product name is too long")
	ProductPriceNotNegative        = errors.New("product price cannot be negative")
	ProductSKUIsRequired           = errors.New("product sku cannot be empty")
	ErrProductNotFound             = errors.New("product not found")
	ProductCategoryIDCannotBeEmpty = errors.New("product category cannot be empty")
	InvalidLimitParameter          = errors.New("invalid limit parameter")
	InvalidPageParameter           = errors.New("invalid page parameter")
)
