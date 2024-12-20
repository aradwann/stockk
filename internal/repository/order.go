package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	internalErrors "stockk/internal/errors"
	"stockk/internal/models"

	"github.com/jackc/pgconn"
)

type OrderRepository interface {
	BeginTransaction() (Transaction, error)
	CreateOrder(ctx context.Context, tx Transaction, order *models.Order) error
	GetOrderByID(ctx context.Context, orderId int) (*models.Order, error)
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

var _ OrderRepository = (*orderRepository)(nil)

func (r *orderRepository) BeginTransaction() (Transaction, error) {
	tx, err := r.db.Begin()
	if err != nil {
		slog.Error("error begin transation", "error", err)
		return nil, internalErrors.Wrap(internalErrors.ErrInternalServer, "transaction failed")
	}
	return tx, nil
}

func (r *orderRepository) CreateOrder(ctx context.Context, tx Transaction, order *models.Order) error {
	query := `
		INSERT INTO orders (created_at)
		VALUES ($1)
		RETURNING id
	`

	var orderID int
	err := tx.QueryRowContext(ctx, query, time.Now()).Scan(&orderID)
	if err != nil {
		slog.Error("failed to create order", "error", err)
		return internalErrors.Wrap(internalErrors.ErrInternalServer, "query failed")
	}

	// Insert order items
	itemQuery := `
		INSERT INTO order_items (order_id, product_id, quantity)
		VALUES ($1, $2, $3)
	`

	for _, item := range order.Items {
		_, err := tx.ExecContext(ctx, itemQuery, orderID, item.ProductID, item.Quantity)
		if err != nil {
			// Check if the error is a PgError
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				// Check for foreign key violation (SQLSTATE 23503)
				if pgErr.Code == "23503" {
					slog.Error("foreign key constraint violation", "detail", pgErr.Detail, "constraint", pgErr.ConstraintName)
					return internalErrors.NewAppError(internalErrors.ErrCodeNotFound, "Resource not found", fmt.Sprintf("Product with ID %d not found", item.ProductID))
				}
			}
			slog.Error("failed to insert order item", "error", err)
			return internalErrors.Wrap(internalErrors.ErrInternalServer, "query failed")
		}
	}

	order.ID = orderID
	return nil
}

func (r *orderRepository) GetOrderByID(ctx context.Context, orderID int) (*models.Order, error) {
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
			return nil, internalErrors.NewAppError(internalErrors.ErrCodeNotFound, "Resource not found", fmt.Sprintf("Order with ID %d not found", orderID))
		}
		slog.Error("failed to retrieve order", "error", err)
		return nil, internalErrors.Wrap(internalErrors.ErrInternalServer, "query failed")
	}

	// Fetch order items
	rows, err := r.db.QueryContext(ctx, itemsQuery, orderID)
	if err != nil {
		slog.Error("failed to retrieve order items", "error", err)
		return nil, internalErrors.Wrap(internalErrors.ErrInternalServer, "query failed")
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			slog.Error("failed to retrieve order item", "error", err)
			return nil, internalErrors.Wrap(internalErrors.ErrInternalServer, "query failed")
		}
		order.Items = append(order.Items, item)
	}

	if err = rows.Err(); err != nil {
		slog.Error("failed to retrieve order item", "error", err)
		return nil, internalErrors.Wrap(internalErrors.ErrInternalServer, "query failed")
	}

	return &order, nil
}
