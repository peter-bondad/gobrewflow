package middleware

import (
	"errors"
	"net/http"

	"gobrewflow/internal/services/categories"
	"gobrewflow/internal/services/invitations"
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
				details[fieldErr.Field()] = fieldErr.Field() + " is required"
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

var statusMap = map[error]int{
	user.InvalidCredentials:                  http.StatusUnauthorized,
	user.ErrInvalidUserRole:                  http.StatusBadRequest,
	products.ErrProductNotFound:              http.StatusNotFound,
	products.ProductNameAlreadyExists:        http.StatusConflict,
	products.ProductIdIsRequired:             http.StatusBadRequest,
	products.ProductInputIsRequired:          http.StatusBadRequest,
	products.ProductNameIsRequired:           http.StatusBadRequest,
	products.ProductSKUIsRequired:            http.StatusBadRequest,
	products.ProductCategoryIDIsRequired:     http.StatusBadRequest,
	products.InvalidLimitParameter:           http.StatusBadRequest,
	products.InvalidPageParameter:            http.StatusBadRequest,
	invitations.ErrInvitationNotFound:        http.StatusNotFound,
	invitations.ErrInvitationAlreadySent:     http.StatusConflict,
	invitations.ErrInvitationLimitReached:    http.StatusForbidden,
	invitations.ErrInvitationAlreadyAccepted: http.StatusConflict,
	invitations.ErrInvitationNotPending:      http.StatusConflict,
	invitations.ErrInvitationExpired:         http.StatusGone,
	invitations.ErrInvitationNotAccepted:     http.StatusConflict,
	invitations.ErrSetupTokenInvalid:         http.StatusNotFound,
	invitations.ErrSetupTokenExpired:         http.StatusGone,
	invitations.ErrPasswordMismatch:          http.StatusBadRequest,
	invitations.ErrEmailAlreadyExists:        http.StatusConflict,
	invitations.ErrForbidden:                 http.StatusForbidden,
	categories.CategoryNameAlreadyExists:     http.StatusConflict,
	categories.ErrCategoryNotFound:           http.StatusNotFound,
	categories.ErrCategoryAlreadyExists:      http.StatusConflict,
	categories.ErrCategoryNotActive:          http.StatusConflict,
	categories.InvalidLimitParameter:         http.StatusBadRequest,
	categories.InvalidPageParameter:          http.StatusBadRequest,
}

func getStatusCode(err error) int {
	for sentinel, code := range statusMap {
		if errors.Is(err, sentinel) {
			return code
		}
	}

	return http.StatusInternalServerError
}
