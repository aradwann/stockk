package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"stockk/internal/models"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) BeginTransaction() (*sql.Tx, error) {
	return r.db.Begin()
}

func (r *OrderRepository) CreateOrder(ctx context.Context, tx *sql.Tx, order *models.Order) error {
	query := `
		INSERT INTO orders (created_at)
		VALUES ($1)
		RETURNING id
	`

	var orderID int
	err := tx.QueryRowContext(ctx, query, time.Now()).Scan(&orderID)
	if err != nil {
		return fmt.Errorf("error creating order: %w", err)
	}

	// Insert order items
	itemQuery := `
		INSERT INTO order_items (order_id, product_id, quantity)
		VALUES ($1, $2, $3)
	`

	for _, item := range order.Items {
		_, err := tx.ExecContext(ctx, itemQuery, orderID, item.ProductID, item.Quantity)
		if err != nil {
			return fmt.Errorf("error creating order item: %w", err)
		}
	}

	order.ID = orderID
	return nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, orderID int) (*models.Order, error) {
	// Main order query
	orderQuery := `
		SELECT id, created_at
		FROM orders
		WHERE id = $1
	`

	// Order items query
	itemsQuery := `
		SELECT product_id, quantity
		FROM order_items
		WHERE order_id = $1
	`

	var order models.Order
	err := r.db.QueryRowContext(ctx, orderQuery, orderID).Scan(
		&order.ID,
		&order.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order with ID %d not found", orderID)
		}
		return nil, fmt.Errorf("error fetching order: %w", err)
	}

	// Fetch order items
	rows, err := r.db.QueryContext(ctx, itemsQuery, orderID)
	if err != nil {
		return nil, fmt.Errorf("error fetching order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			return nil, fmt.Errorf("error scanning order item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading order items: %w", err)
	}

	return &order, nil
}

func (r *OrderRepository) ListOrders(ctx context.Context, limit, offset int) ([]models.Order, error) {
	query := `
		SELECT id, created_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning order: %w", err)
		}

		// Fetch items for each order
		itemsQuery := `
			SELECT product_id, quantity
			FROM order_items
			WHERE order_id = $1
		`
		itemRows, err := r.db.QueryContext(ctx, itemsQuery, order.ID)
		if err != nil {
			return nil, fmt.Errorf("error fetching order items: %w", err)
		}
		defer itemRows.Close()

		for itemRows.Next() {
			var item models.OrderItem
			if err := itemRows.Scan(&item.ProductID, &item.Quantity); err != nil {
				return nil, fmt.Errorf("error scanning order item: %w", err)
			}
			order.Items = append(order.Items, item)
		}

		if err = itemRows.Err(); err != nil {
			return nil, fmt.Errorf("error reading order items: %w", err)
		}

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading orders: %w", err)
	}

	return orders, nil
}
