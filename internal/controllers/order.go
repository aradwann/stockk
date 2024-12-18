package controllers

import (
	"encoding/json"
	"net/http"

	"log/slog"

	"stockk/internal/models"
	"stockk/internal/service"
)

type OrderController struct {
	orderService      *service.OrderService
	ingredientService *service.IngredientService
}

func NewOrderController(
	orderService *service.OrderService,
	ingredientService *service.IngredientService,
) *OrderController {
	return &OrderController{
		orderService:      orderService,
		ingredientService: ingredientService,
	}
}

func (oc *OrderController) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var orderRequest struct {
		Products []models.OrderItem `json:"products"`
	}

	if err := json.NewDecoder(r.Body).Decode(&orderRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Create order and update ingredient stocks
	order, err := oc.orderService.CreateOrder(r.Context(), orderRequest.Products)
	if err != nil {
		slog.Error("Failed to create order", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Check ingredient levels and send alerts if needed

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}
