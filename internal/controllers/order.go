package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"log/slog"

	"stockk/internal/models"
	"stockk/internal/service"
	"stockk/internal/validator"

	"github.com/go-chi/render"
)

type OrderController struct {
	orderService *service.OrderService
}

func NewOrderController(
	orderService *service.OrderService,
) *OrderController {
	return &OrderController{
		orderService: orderService,
	}
}

type orderRequest struct {
	Products []models.OrderItem `json:"products"`
}

func (oc *OrderController) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var orderRequest orderRequest
	if err := json.NewDecoder(r.Body).Decode(&orderRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	err := validateCreateOrderRequest(&orderRequest)
	if err != nil {
		slog.Error("Failed to create order", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, order)
}

func validateCreateOrderRequest(orderReq *orderRequest) error {
	if orderReq == nil {
		return errors.New("error parsing request body")
	}
	for _, p := range orderReq.Products {
		if err := validator.ValidateID(p.ProductID); err != nil {
			return err
		}
		if err := validator.ValidateQuantity(p.Quantity); err != nil {
			return err
		}
	}
	return nil

}
