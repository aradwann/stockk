package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	internalErrors "stockk/internal/errors"
)

type ErrorResponse struct {
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

// handleServiceError categorizes and responds to errors from the service layer.
func handleServiceError(w http.ResponseWriter, err error) {
	var appErr *internalErrors.AppError
	if errors.As(err, &appErr) {
		// Extract the structured error details
		response := ErrorResponse{
			Message: appErr.Message,
			Code:    appErr.Code,
			Details: appErr.Details, // Optional
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(appErr.Code)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Fallback for unknown errors
	response := ErrorResponse{
		Message: "Internal server error",
		Code:    http.StatusInternalServerError,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(response)
}
