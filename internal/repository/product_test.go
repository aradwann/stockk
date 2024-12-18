package repository

import (
	"context"
	"stockk/internal/models"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestProductRepository_GetByID(t *testing.T) {
	// Create a mock DB and mock objects
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer db.Close()

	// Create the repository
	repo := NewProductRepository(db)

	// Define the product ID to test
	productID := 1

	// Expected product and ingredients
	expectedProduct := models.Product{
		ID:   productID,
		Name: "Burger",
		Ingredients: []models.ProductIngredient{
			{ProductID: productID, IngredientID: 1, Amount: 150},
			{ProductID: productID, IngredientID: 2, Amount: 30},
			{ProductID: productID, IngredientID: 3, Amount: 20},
		},
	}

	// Mock the product query
	mock.ExpectQuery(`SELECT id, name FROM products WHERE id = \$1`).
		WithArgs(productID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(productID, "Burger"))

	// Mock the product ingredients query
	mock.ExpectQuery(`SELECT pi.product_id, pi.ingredient_id, pi.amount`).
		WithArgs(productID).
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "ingredient_id", "amount"}).
			AddRow(productID, 1, 150).
			AddRow(productID, 2, 30).
			AddRow(productID, 3, 20))

	// Call the method under test
	product, err := repo.GetByID(context.Background(), nil, productID)

	// Assertions
	assert.NoError(t, err, "Expected no error, got %v", err)
	assert.NotNil(t, product, "Expected product to be returned")
	assert.Equal(t, expectedProduct, *product, "Expected and actual products do not match")
}

func TestProductRepository_GetByID_NotFound(t *testing.T) {
	// Create a mock DB and mock objects
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer db.Close()

	// Create the repository
	repo := NewProductRepository(db)

	// Define the product ID to test
	productID := 999

	// Mock the product query to return no rows
	mock.ExpectQuery(`SELECT id, name FROM products WHERE id = \$1`).
		WithArgs(productID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	// Call the method under test
	product, err := repo.GetByID(context.Background(), nil, productID)

	// Assertions
	assert.Error(t, err, "Expected error when product not found")
	assert.Nil(t, product, "Expected no product to be returned")
}
