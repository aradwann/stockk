package errors

import "fmt"

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	return fmt.Sprintf("code: %d, message: %s, error: %v", e.Code, e.Message, e.Err)
}

func NewAppError(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Predefined error categories
var (
	ErrNotFound          = NewAppError(404, "resource not found", nil)
	ErrInternalServer    = NewAppError(500, "internal server error", nil)
	ErrValidation        = NewAppError(400, "validation failed", nil)
	ErrInsufficientStock = NewAppError(400, "insufficient ingredient stock", nil)
)
