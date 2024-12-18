package repository

import (
	"context"
	"stockk/internal/models"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestOrderRepository_CreateOrder(t *testing.T) {
	// Create a mock DB and mock objects
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer db.Close()

	// Create the repository
	repo := NewOrderRepository(db)

	// Create an order to insert
	order := &models.Order{
		Items: []models.OrderItem{
			{ProductID: 1, Quantity: 2},
			{ProductID: 2, Quantity: 1},
		},
	}

	// Mock the transaction and start expectations
	mock.ExpectBegin() // Expect the transaction to start

	// Mock the query for creating an order and returning its ID
	mock.ExpectQuery(`^INSERT INTO orders \(created_at\) VALUES \(\$1\) RETURNING id$`).
		WithArgs(sqlmock.AnyArg()).                               // Accept any value for the time argument
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)) // Mock return of order ID = 1

	// Mock the insertion of order items
	mock.ExpectExec(`^INSERT INTO order_items \(order_id, product_id, quantity\) VALUES \(\$1, \$2, \$3\)$`).
		WithArgs(1, 1, 2).
		WillReturnResult(sqlmock.NewResult(1, 1)) // Mock success for this insert

	mock.ExpectExec(`^INSERT INTO order_items \(order_id, product_id, quantity\) VALUES \(\$1, \$2, \$3\)$`).
		WithArgs(1, 2, 1).
		WillReturnResult(sqlmock.NewResult(1, 1)) // Mock success for this insert

	// Expect the transaction to commit at the end
	mock.ExpectCommit()

	// Begin a transaction (this will match the mock expectation)
	tx, err := repo.BeginTransaction()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Call the method under test
	err = repo.CreateOrder(context.Background(), tx, order)

	// Assertions
	assert.NoError(t, err, "Expected no error, got %v", err)
	assert.Equal(t, 1, order.ID, "Expected order ID to be 1")

	// Commit the transaction after the operation
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Ensure that the mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %v", err)
	}
}

func TestOrderRepository_GetOrderByID(t *testing.T) {
	// Create a mock DB and mock objects
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer db.Close()

	// Create the repository
	repo := NewOrderRepository(db)

	// Order ID to search
	orderID := 1

	// Expected order and order items
	expectedOrder := &models.Order{
		ID:        orderID,
		CreatedAt: time.Now(),
		Items: []models.OrderItem{
			{ProductID: 1, Quantity: 2},
			{ProductID: 2, Quantity: 1},
		},
	}

	// Mock the query for the order
	mock.ExpectQuery(`SELECT id, created_at FROM orders WHERE id = \$1`).
		WithArgs(orderID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(orderID, expectedOrder.CreatedAt))

	// Mock the query for order items
	mock.ExpectQuery(`SELECT product_id, quantity FROM order_items WHERE order_id = \$1`).
		WithArgs(orderID).
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "quantity"}).
			AddRow(1, 2).
			AddRow(2, 1))

	// Call the method under test
	order, err := repo.GetOrderByID(context.Background(), orderID)

	// Assertions
	assert.NoError(t, err, "Expected no error, got %v", err)
	assert.NotNil(t, order, "Expected order to be returned")
	assert.Equal(t, expectedOrder, order, "Expected and actual orders do not match")
}

func TestOrderRepository_GetOrderByID_NotFound(t *testing.T) {
	// Create a mock DB and mock objects
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer db.Close()

	// Create the repository
	repo := NewOrderRepository(db)

	// Order ID to search
	orderID := 999

	// Mock the query for the order
	mock.ExpectQuery(`SELECT id, created_at FROM orders WHERE id = \$1`).
		WithArgs(orderID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}))

	// Call the method under test
	order, err := repo.GetOrderByID(context.Background(), orderID)

	// Assertions
	assert.Error(t, err, "Expected error when order is not found")
	assert.Nil(t, order, "Expected no order to be returned")
}

func TestOrderRepository_ListOrders(t *testing.T) {
	// Create a mock DB and mock objects
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer db.Close()

	// Create the repository
	repo := NewOrderRepository(db)

	// Mock the query for the orders list
	mock.ExpectQuery(`SELECT id, created_at FROM orders ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow(1, time.Now()).
			AddRow(2, time.Now().Add(-1*time.Hour)))

	// Mock the query for order items
	mock.ExpectQuery(`SELECT product_id, quantity FROM order_items WHERE order_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "quantity"}).
			AddRow(1, 2).
			AddRow(2, 1))

	mock.ExpectQuery(`SELECT product_id, quantity FROM order_items WHERE order_id = \$1`).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "quantity"}).
			AddRow(1, 3))

	// Call the method under test
	orders, err := repo.ListOrders(context.Background(), 10, 0)

	// Assertions
	assert.NoError(t, err, "Expected no error, got %v", err)
	assert.Len(t, orders, 2, "Expected 2 orders to be returned")
}
