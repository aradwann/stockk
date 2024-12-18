package service

import (
	"context"
	"stockk/internal/models"
	"stockk/internal/repository"
	"time"
)

type OrderService struct {
	orderRepo         *repository.OrderRepository
	productRepo       *repository.ProductRepository
	ingredientService *IngredientService
}

func NewOrderService(orderRepo *repository.OrderRepository) *OrderService {
	return &OrderService{orderRepo: orderRepo}
}

func (os *OrderService) CreateOrder(ctx context.Context, orderItems []models.OrderItem) (*models.Order, error) {
	// Begin transaction
	tx, err := os.orderRepo.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create order
	order := &models.Order{
		Items:     orderItems,
		CreatedAt: time.Now(),
	}

	if err := os.orderRepo.CreateOrder(ctx, tx, order); err != nil {
		return nil, err
	}

	// Process each product and update ingredient stocks
	// for _, item := range orderItems {
	// 	product, err := os.productRepo.GetByID(ctx, item.ProductID)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// for _, ingredient := range product.Ingredients {
	// TODO: get ingredient and update stock value
	// Reduce stock based on product quantity
	// ingredient.CurrentStock -= (ingredient.Amount * float64(item.Quantity))
	// }

	// Update ingredient stocks
	// if err := os.ingredientService.UpdateIngredientStock(ctx, product.Ingredients); err != nil {
	// 	return nil, err
	// }
	// }

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return order, nil
}
