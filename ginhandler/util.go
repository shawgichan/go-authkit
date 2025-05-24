package ginhandler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shawgichan/go-authkit/core" // Adjust import path
)

// ErrorResponse is a generic JSON error response.
type ErrorResponse struct {
	Status  string      `json:"status"`
	Code    string      `json:"code"` // SDK-defined error code or HTTP status text
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// RespondWithError sends a JSON error response.
func RespondWithError(c *gin.Context, httpStatusCode int, sdkErrorCode string, message string, details interface{}) {
	// You might want to log the error server-side here too
	c.AbortWithStatusJSON(httpStatusCode, ErrorResponse{
		Status:  "error",
		Code:    sdkErrorCode,
		Message: message,
		Details: details,
	})
}

// MapSDKErrorToHTTP maps core SDK errors to HTTP status codes and error details.
func MapSDKErrorToHTTP(c *gin.Context, err error) {
	// Default error
	httpStatus := http.StatusInternalServerError
	errCode := "INTERNAL_ERROR"
	errMsg := "An unexpected error occurred"
	var errDetails interface{}

	if err != nil {
		errMsg = err.Error() // Default message from error
	}

	switch {
	case errors.Is(err, core.ErrNotFound):
		httpStatus = http.StatusNotFound
		errCode = "NOT_FOUND"
	case errors.Is(err, core.ErrDuplicateEmail), errors.Is(err, core.ErrDuplicateUsername):
		httpStatus = http.StatusConflict
		errCode = "CONFLICT_RESOURCE"
	case errors.Is(err, core.ErrInvalidCredentials):
		httpStatus = http.StatusUnauthorized
		errCode = "INVALID_CREDENTIALS"
	case errors.Is(err, core.ErrUserNotVerified):
		httpStatus = http.StatusForbidden // Or StatusUnauthorized with specific code
		errCode = "USER_NOT_VERIFIED"
	case errors.Is(err, core.ErrTokenInvalid), errors.Is(err, core.ErrTokenExpired):
		httpStatus = http.StatusUnauthorized
		errCode = "INVALID_TOKEN"
	case errors.Is(err, core.ErrForbidden):
		httpStatus = http.StatusForbidden
		errCode = "FORBIDDEN"
	// Add more mappings as needed
	default:
		// Check for validation errors if you use a library like validator
		// var verr validator.ValidationErrors
		// if errors.As(err, &verr) {
		//     httpStatus = http.StatusBadRequest
		//     errCode = "VALIDATION_ERROR"
		//     errDetails = formatValidationErrors(verr) // You'd write this helper
		// }
		// For a simple start, you might not have this complex validation error handling.
		if err != nil { // If it's some other known error type, use its message
			errMsg = err.Error()
		}
	}
	RespondWithError(c, httpStatus, errCode, errMsg, errDetails)
}

// SuccessResponse is a generic JSON success response.
type SuccessResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
}

// RespondWithSuccess sends a JSON success response.
func RespondWithSuccess(c *gin.Context, httpStatusCode int, data interface{}) {
	c.JSON(httpStatusCode, SuccessResponse{
		Status: "success",
		Data:   data,
	})
}
