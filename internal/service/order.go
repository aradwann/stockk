package service

import (
	"context"
	"database/sql"
	"fmt"
	"stockk/internal/models"
	"stockk/internal/repository"
	"time"
)

type OrderService struct {
	orderRepo      *repository.OrderRepository
	productRepo    *repository.ProductRepository
	ingredientRepo *repository.IngredientRepository
}

func NewOrderService(orderRepo *repository.OrderRepository, productRepo *repository.ProductRepository, ingredientRepo *repository.IngredientRepository) *OrderService {
	return &OrderService{orderRepo: orderRepo}
}

// TODO: handle the following cases:
// - ingredients amounts are zero or negative
func (os *OrderService) CreateOrder(ctx context.Context, orderItems []models.OrderItem) (*models.Order, error) {
	// Begin transaction
	tx, err := os.orderRepo.BeginTransaction()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create the order
	order := &models.Order{
		Items:     orderItems,
		CreatedAt: time.Now(),
	}

	if err := os.orderRepo.CreateOrder(ctx, tx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Process each product and update ingredient stocks
	for _, item := range orderItems {
		if err := os.processOrderItem(ctx, tx, item); err != nil {
			// TODO: don't return error details ,return a generic error and log the error details
			return nil, fmt.Errorf("failed to process order item (ProductID: %d): %w", item.ProductID, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return order, nil
}

func (os *OrderService) processOrderItem(ctx context.Context, tx *sql.Tx, item models.OrderItem) error {
	// Retrieve the product
	product, err := os.productRepo.GetByID(ctx, tx, item.ProductID)
	if err != nil {
		return fmt.Errorf("failed to retrieve product: %w", err)
	}

	// Update stock for each ingredient
	for _, productIngredient := range product.Ingredients {
		if err := os.updateIngredientStock(ctx, tx, productIngredient.IngredientID, productIngredient.Amount, item.Quantity); err != nil {
			return fmt.Errorf("failed to update stock for ingredient %d: %w", productIngredient.IngredientID, err)
		}
	}

	return nil
}

func (os *OrderService) updateIngredientStock(ctx context.Context, tx *sql.Tx, ingredientID int, amountPerUnit float64, quantity int) error {
	// Retrieve the ingredient
	ingredient, err := os.ingredientRepo.GetIngredientByID(ctx, tx, ingredientID)
	if err != nil {
		return fmt.Errorf("failed to retrieve ingredient: %w", err)
	}

	// Calculate and update stock
	newStock := ingredient.CurrentStock - (amountPerUnit * float64(quantity))
	if err := os.ingredientRepo.UpdateStock(ctx, tx, ingredientID, newStock); err != nil {
		return fmt.Errorf("failed to update ingredient stock: %w", err)
	}

	return nil
}
