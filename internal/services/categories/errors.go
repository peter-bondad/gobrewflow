package categories

import "errors"

var (
	CategoryNameAlreadyExists = errors.New("category already exists")
	ErrCategoryNotFound       = errors.New("category not found")
	ErrCategoryAlreadyExists  = errors.New("category already exists")
	ErrCategoryNotActive      = errors.New("category is not active")
)
