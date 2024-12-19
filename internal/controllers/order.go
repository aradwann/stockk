package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"log/slog"

	internalErrors "stockk/internal/errors"
	"stockk/internal/models"
	"stockk/internal/service"
	"stockk/internal/validator"

	"github.com/go-chi/render"
)

type OrderController struct {
	orderService      service.OrderService
	ingredientService service.IngredientService
}

func NewOrderController(
	orderService service.OrderService,
	ingredientService service.IngredientService,
) *OrderController {
	return &OrderController{
		orderService:      orderService,
		ingredientService: ingredientService,
	}
}

type orderRequest struct {
	Products []models.OrderItem `json:"products"`
}

func (oc *OrderController) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var orderRequest orderRequest

	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(&orderRequest); err != nil {
		slog.Error("Invalid request payload", "error", err)
		handleServiceError(w, errors.New("Invalid request payload"))
		return
	}

	// Validate the request
	if err := validateCreateOrderRequest(&orderRequest); err != nil {
		slog.Error("Request validation failed", "error", err)
		handleServiceError(w, err)
		return
	}

	// Create order and update ingredient stocks
	order, err := oc.orderService.CreateOrder(r.Context(), orderRequest.Products)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	// Check ingredient levels and send alerts if necessary
	if err := oc.ingredientService.CheckIngredientLevelsAndAlert(r.Context()); err != nil {
		slog.Warn("Failed to send ingredient alert email", "error", err)
	}

	// Respond with the created order
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, order)
}

// validateCreateOrderRequest validates the incoming order request.
func validateCreateOrderRequest(orderReq *orderRequest) error {
	if orderReq == nil {
		return errors.New("error parsing request body")
	}

	for _, p := range orderReq.Products {
		if err := validator.ValidateID(p.ProductID); err != nil {
			return internalErrors.Wrap(internalErrors.ErrValidation, "invalid product ID: "+err.Error())

		}
		if err := validator.ValidateQuantity(p.Quantity); err != nil {
			return internalErrors.Wrap(internalErrors.ErrValidation, "invalid product quantity: "+err.Error())

		}
	}

	return nil
}
