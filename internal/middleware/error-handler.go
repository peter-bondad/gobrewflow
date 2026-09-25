package middleware

import (
	"errors"
	"net/http"
	"unicode"

	"gobrewflow/internal/services/categories"
	"gobrewflow/internal/services/inventory"
	"gobrewflow/internal/services/inventory_movements"
	"gobrewflow/internal/services/invitations"
	"gobrewflow/internal/services/order_items"
	"gobrewflow/internal/services/orders"
	"gobrewflow/internal/services/products"
	"gobrewflow/internal/services/user"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
			details := make(map[string]string)
			for _, fieldErr := range validationErrors {
				field := toSnakeCase(fieldErr.Field())

				switch fieldErr.Tag() {
				case "required":
					details[field] = field + " is required"

				case "gt":
					details[field] = field + " must be greater than 0"

				case "gte":
					details[field] = field + " must be greater than or equal to 0"

				case "lt":
					details[field] = field + " must be less than " + fieldErr.Param()

				case "lte":
					details[field] = field + " must be less than or equal to " + fieldErr.Param()

				default:
					details[field] = field + " is invalid"
				}
			}
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: ErrorDetail{
					Code:    "VALIDATION_ERROR",
					Message: "validation failed",
					Details: details,
				},
			})
			return
		}

		c.JSON(getStatusCode(err), ErrorResponse{
			Error: ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
	}
}

func toSnakeCase(s string) string {
	var result []rune

	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '_')
		}

		result = append(result, unicode.ToLower(r))
	}

	return string(result)
}

var statusMap = map[error]int{
	user.InvalidCredentials:                     http.StatusUnauthorized,
	user.ErrInvalidUserRole:                     http.StatusBadRequest,
	products.ErrProductNotFound:                 http.StatusNotFound,
	products.ProductNameAlreadyExists:           http.StatusConflict,
	products.ProductIdIsRequired:                http.StatusBadRequest,
	products.ProductInputIsRequired:             http.StatusBadRequest,
	products.ProductNameCannotBeEmpty:           http.StatusBadRequest,
	products.ProductNameTooLong:                 http.StatusBadRequest,
	products.ProductPriceNotNegative:            http.StatusBadRequest,
	products.ProductSKUIsRequired:               http.StatusBadRequest,
	products.ProductCategoryIDCannotBeEmpty:     http.StatusBadRequest,
	products.InvalidLimitParameter:              http.StatusBadRequest,
	products.InvalidPageParameter:               http.StatusBadRequest,
	products.ErrInvalidProductID:                http.StatusBadRequest,
	products.ErrInvalidUnitPrice:                http.StatusBadRequest,
	products.ErrNoProducts:                      http.StatusBadRequest,
	invitations.ErrInvitationNotFound:           http.StatusNotFound,
	invitations.ErrInvitationAlreadySent:        http.StatusConflict,
	invitations.ErrInvitationLimitReached:       http.StatusForbidden,
	invitations.ErrInvitationAlreadyAccepted:    http.StatusConflict,
	invitations.ErrInvitationNotPending:         http.StatusConflict,
	invitations.ErrInvitationExpired:            http.StatusGone,
	invitations.ErrInvitationNotAccepted:        http.StatusConflict,
	invitations.ErrSetupTokenInvalid:            http.StatusNotFound,
	invitations.ErrSetupTokenExpired:            http.StatusGone,
	invitations.ErrPasswordMismatch:             http.StatusBadRequest,
	invitations.ErrEmailAlreadyExists:           http.StatusConflict,
	invitations.ErrForbidden:                    http.StatusForbidden,
	categories.CategoryNameAlreadyExists:        http.StatusConflict,
	categories.ErrCategoryNotFound:              http.StatusNotFound,
	categories.ErrCategoryAlreadyExists:         http.StatusConflict,
	categories.ErrCategoryNotActive:             http.StatusConflict,
	categories.InvalidLimitParameter:            http.StatusBadRequest,
	categories.InvalidPageParameter:             http.StatusBadRequest,
	inventory.ErrInventoryNotFound:              http.StatusNotFound,
	inventory.ErrInvalidQuantity:                http.StatusBadRequest,
	inventory.ErrInsufficientStock:              http.StatusConflict,
	orders.ErrInvalidOrderID:                    http.StatusBadRequest,
	orders.ErrInvalidCashierID:                  http.StatusBadRequest,
	order_items.ErrInvalidOrderID:               http.StatusBadRequest,
	inventory_movements.ErrInvalidMovementType:  http.StatusBadRequest,
	inventory_movements.ErrInvalidQuantity:      http.StatusBadRequest,
	inventory_movements.ErrProductNotFound:      http.StatusNotFound,
	inventory_movements.ErrInventoryNotFound:    http.StatusNotFound,
	inventory_movements.ErrInsufficientStock:    http.StatusConflict,
	inventory_movements.ErrMovementNotFound:     http.StatusNotFound,
	inventory_movements.InvalidLimitParameter:   http.StatusBadRequest,
	inventory_movements.InvalidPageParameter:    http.StatusBadRequest,
	inventory_movements.ErrInvalidStockIncrease: http.StatusConflict,
}

func getStatusCode(err error) int {
	for sentinel, code := range statusMap {
		if errors.Is(err, sentinel) {
			return code
		}
	}

	if _, ok := errors.AsType[*orders.InsufficientStockError](err); ok {
		return http.StatusConflict
	}

	return http.StatusInternalServerError
}
