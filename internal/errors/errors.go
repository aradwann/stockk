package errors

import "fmt"

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Predefined error categories
var (
	ErrNotFound          = NewAppError(404, "resource not found")
	ErrInternalServer    = NewAppError(500, "internal server error")
	ErrValidation        = NewAppError(400, "validation failed")
	ErrInsufficientStock = NewAppError(409, "insufficient ingredient stock")
)
