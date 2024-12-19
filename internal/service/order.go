package service

import (
	"context"
	"database/sql"
	"errors"
	internalErrors "stockk/internal/errors"
	"stockk/internal/models"
	"stockk/internal/repository"
	"time"
)

type OrderService interface {
	CreateOrder(ctx context.Context, orderItems []models.OrderItem) (*models.Order, error)
}

type orderService struct {
	orderRepo      repository.OrderRepository
	productRepo    repository.ProductRepository
	ingredientRepo repository.IngredientRepository
}

func NewOrderService(orderRepo repository.OrderRepository, productRepo repository.ProductRepository, ingredientRepo repository.IngredientRepository) OrderService {
	return &orderService{
		orderRepo:      orderRepo,
		productRepo:    productRepo,
		ingredientRepo: ingredientRepo,
	}
}

var _ OrderService = (*orderService)(nil)

func (os *orderService) CreateOrder(ctx context.Context, orderItems []models.OrderItem) (*models.Order, error) {
	// Begin transaction
	tx, err := os.orderRepo.BeginTransaction()
	if err != nil {
		return nil, internalErrors.Wrap(internalErrors.ErrInternalServer, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Create the order
	order := &models.Order{
		Items:     orderItems,
		CreatedAt: time.Now(),
	}

	if err := os.orderRepo.CreateOrder(ctx, tx, order); err != nil {
		if errors.Is(err, internalErrors.ErrNotFound) {
			return nil, err
		} else {
			return nil, internalErrors.Wrap(internalErrors.ErrInternalServer, "failed to create order")
		}
	}

	// Process each product and update ingredient stocks
	for _, item := range orderItems {
		if err := os.processOrderItem(ctx, tx, item); err != nil {
			return nil, err
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		if errors.Is(err, internalErrors.ErrNotFound) {
			return nil, err
		} else {
			return nil, internalErrors.Wrap(internalErrors.ErrInternalServer, "failed to commit transaction")
		}
	}

	return order, nil
}

func (os *orderService) processOrderItem(ctx context.Context, tx *sql.Tx, item models.OrderItem) error {
	// Retrieve the product
	product, err := os.productRepo.GetProductById(ctx, tx, item.ProductID)
	if err != nil {
		if errors.Is(err, internalErrors.ErrNotFound) {
			return err
		} else {
			return internalErrors.Wrap(internalErrors.ErrInternalServer, "failed to get product")
		}
	}

	// Update stock for each ingredient
	for _, productIngredient := range product.Ingredients {
		if err := os.updateIngredientStock(ctx, tx, productIngredient.IngredientID, productIngredient.Amount, item.Quantity); err != nil {
			return err
		}
	}

	return nil
}

func (os *orderService) updateIngredientStock(ctx context.Context, tx *sql.Tx, ingredientID int, amountPerUnit float64, quantity int) error {
	// Retrieve the ingredient
	ingredient, err := os.ingredientRepo.GetIngredientByID(ctx, tx, ingredientID)
	if err != nil {
		if errors.Is(err, internalErrors.ErrNotFound) {
			return err
		} else {
			return internalErrors.Wrap(internalErrors.ErrInternalServer, "failed to get ingredient stock")
		}
	}

	// Calculate and update stock
	newStock := ingredient.CurrentStock - (amountPerUnit * float64(quantity))
	// validate that remaining stock is positive number
	if newStock < 0 {
		return internalErrors.ErrInsufficientStock
	}

	if err := os.ingredientRepo.UpdateStock(ctx, tx, ingredientID, newStock); err != nil {
		if errors.Is(err, internalErrors.ErrNotFound) {
			return err
		} else {
			return internalErrors.Wrap(internalErrors.ErrInternalServer, "failed to update ingredient stock")
		}
	}

	return nil
}
