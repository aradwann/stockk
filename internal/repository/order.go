package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"stockk/internal/errors"
	"stockk/internal/models"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) BeginTransaction() (*sql.Tx, error) {
	tx, err := r.db.Begin()
	if err != nil {
		slog.Error("error begin transation", "error", err)
		return nil, errors.Wrap(errors.ErrInternalServer, "transaction failed")
	}
	return tx, nil
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
		slog.Error("failed to create order", "error", err)
		return errors.Wrap(errors.ErrInternalServer, "query failed")
	}

	// Insert order items
	itemQuery := `
		INSERT INTO order_items (order_id, product_id, quantity)
		VALUES ($1, $2, $3)
	`

	for _, item := range order.Items {
		_, err := tx.ExecContext(ctx, itemQuery, orderID, item.ProductID, item.Quantity)
		if err != nil {
			slog.Error("failed to create order", "error", err)
			return errors.Wrap(errors.ErrInternalServer, "query failed")
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
			return nil, errors.Wrap(errors.ErrNotFound, "order not found")
		}
		slog.Error("failed to retrieve order", "error", err)
		return nil, errors.Wrap(errors.ErrInternalServer, "query failed")
	}

	// Fetch order items
	rows, err := r.db.QueryContext(ctx, itemsQuery, orderID)
	if err != nil {
		appErr := errors.NewAppError(http.StatusInternalServerError, "error fetching order items", err)
		slog.Error(appErr.Message, "orderID", orderID, "error", err)
		return nil, appErr
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			slog.Error("failed to retrieve order item", "error", err)
			return nil, errors.Wrap(errors.ErrInternalServer, "query failed")
		}
		order.Items = append(order.Items, item)
	}

	if err = rows.Err(); err != nil {
		slog.Error("failed to retrieve order item", "error", err)
		return nil, errors.Wrap(errors.ErrInternalServer, "query failed")
	}

	return &order, nil
}
