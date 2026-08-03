package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"todo-api/internal/dto"
)

// validate is a package-level validator instance (reused for performance)
var validate = validator.New()

// respondSuccess writes a successful JSON response
func respondSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	resp := dto.SuccessResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	}
	c.JSON(statusCode, resp)
}

// respondError writes an error JSON response
func respondError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, dto.ErrorResponse{
		Status:  "error",
		Message: message,
	})
}

// respondValidationError writes a validation error response with field-level details
func respondValidationError(c *gin.Context, err error) {
	var fieldErrors []dto.FieldError

	var ve validator.ValidationErrors
	if ok := parseValidationErrors(err, &ve); ok {
		for _, fe := range ve {
			fieldErrors = append(fieldErrors, dto.FieldError{
				Field:   strings.ToLower(fe.Field()),
				Message: validationMessage(fe),
			})
		}
	}

	c.JSON(http.StatusBadRequest, dto.ErrorResponse{
		Status:  "error",
		Message: "Validation failed",
		Errors:  fieldErrors,
	})
}

// parseValidationErrors attempts to cast the error to validator.ValidationErrors
func parseValidationErrors(err error, target *validator.ValidationErrors) bool {
	if ve, ok := err.(validator.ValidationErrors); ok {
		*target = ve
		return true
	}
	return false
}

// validationMessage returns a human-readable message for a validation error
func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "min":
		return fe.Field() + " must be at least " + fe.Param() + " characters"
	case "max":
		return fe.Field() + " must be at most " + fe.Param() + " characters"
	case "email":
		return "must be a valid email address"
	case "oneof":
		return fe.Field() + " must be one of: " + fe.Param()
	case "datetime":
		return fe.Field() + " must be in format: YYYY-MM-DD"
	default:
		return fe.Field() + " is invalid"
	}
}
