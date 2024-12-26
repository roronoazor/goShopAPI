package libs

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

func NewValidationError(err error) ErrorResponse {
	var validationErrors []ValidationError

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			var errorMsg string
			switch e.Tag() {
			case "required":
				errorMsg = fmt.Sprintf("%s is required", e.Field())
			case "email":
				errorMsg = "Must be a valid email address"
			case "min":
				errorMsg = fmt.Sprintf("%s must be at least %s characters long", e.Field(), e.Param())
			default:
				errorMsg = fmt.Sprintf("%s is invalid", e.Field())
			}

			validationErrors = append(validationErrors, ValidationError{
				Field:   strings.ToLower(e.Field()),
				Message: errorMsg,
			})
		}
	}

	return ErrorResponse{
		Status:  "error",
		Message: "Validation failed",
		Errors:  validationErrors,
	}
}
