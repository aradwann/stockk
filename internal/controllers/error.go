package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	internalErrors "stockk/internal/errors"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"` // Optional additional details
}

// handleServiceError categorizes and responds to errors from the service layer.
func handleServiceError(w http.ResponseWriter, err error) {
	var statusCode int
	var message string

	switch {
	case errors.Is(err, internalErrors.ErrNotFound):
		statusCode = http.StatusNotFound
		message = "Resource not found"
	case errors.Is(err, internalErrors.ErrValidation):
		statusCode = http.StatusBadRequest
		message = "Validation error"
	case errors.Is(err, internalErrors.ErrInsufficientStock):
		statusCode = http.StatusConflict
		message = "Insufficient stock"
	default:
		slog.Error("Internal server error", "error", err)
		statusCode = http.StatusInternalServerError
		message = "Internal server error"
	}

	// Create and render the JSON error response
	errorResponse := ErrorResponse{
		Message: message,
		Code:    statusCode,
		Details: err.Error(), // Optionally include the original error message
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}
