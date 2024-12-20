package errors

import (
	"errors"
	"fmt"
)

// AppError represents a structured application error.
type AppError struct {
	Code    int    // HTTP-like error code
	Message string // Human-readable error message
	Details string // Optional additional details
}

func (e *AppError) Error() string {
	return e.Message
}

// NewAppError creates a new instance of AppError.
func NewAppError(code int, message string, details ...string) *AppError {
	detailMessage := ""
	if len(details) > 0 {
		detailMessage = details[0]
	}
	return &AppError{
		Code:    code,
		Message: message,
		Details: detailMessage,
	}
}

// Wrap wraps an existing error with additional context and returns an AppError.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Unwrap extracts the original error if available (optional for compatibility).
func (e *AppError) Unwrap() error {
	return errors.New(e.Details)
}

// Predefined error categories
var (
	ErrNotFound          = NewAppError(ErrCodeNotFound, "resource not found")
	ErrInternalServer    = NewAppError(ErrCodeInternalServer, "internal server error")
	ErrValidation        = NewAppError(ErrCodeValidation, "validation failed")
	ErrInsufficientStock = NewAppError(ErrCodeInsufficientStock, "insufficient ingredient stock")
)

// Predefined error codes
const (
	ErrCodeNotFound          = 404
	ErrCodeInternalServer    = 500
	ErrCodeValidation        = 400
	ErrCodeInsufficientStock = 409
)
