package repository

import (
	"context"
	"database/sql"
	"testing"

	"stockk/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestIngredientRepository_GetIngredientByID(t *testing.T) {
	// Create a mock DB and mock objects
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer db.Close()

	repo := NewIngredientRepository(db)

	// Define the expected ingredient data
	ingredientID := 1
	expectedIngredient := &models.Ingredient{
		ID:           ingredientID,
		Name:         "Sugar",
		TotalStock:   100,
		CurrentStock: 40,
		AlertSent:    false,
	}

	// Mock the query for getting an ingredient by ID
	mock.ExpectQuery(`SELECT id, name, total_stock, current_stock, alert_sent FROM ingredients WHERE id = \$(\d)`).
		WithArgs(ingredientID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "total_stock", "current_stock", "alert_sent"}).
			AddRow(expectedIngredient.ID, expectedIngredient.Name, expectedIngredient.TotalStock, expectedIngredient.CurrentStock, expectedIngredient.AlertSent))

	// Call the method under test
	ingredient, err := repo.GetIngredientByID(context.Background(), nil, ingredientID)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedIngredient, ingredient)

	// Ensure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %v", err)
	}
}

func TestIngredientRepository_GetIngredientByID_NotFound(t *testing.T) {
	// Create a mock DB and mock objects
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer db.Close()

	repo := NewIngredientRepository(db)

	// Define the expected ingredient ID that does not exist
	ingredientID := 999

	// Mock the query for getting an ingredient by ID, returning no rows
	mock.ExpectQuery(`SELECT id, name, total_stock, current_stock, alert_sent FROM ingredients WHERE id = \$(\d)`).
		WithArgs(ingredientID).
		WillReturnError(sql.ErrNoRows)

	// Call the method under test
	ingredient, err := repo.GetIngredientByID(context.Background(), nil, ingredientID)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, ingredient)
	assert.Contains(t, err.Error(), "Resource not found")

	// Ensure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %v", err)
	}
}

func TestIngredientRepository_UpdateStock(t *testing.T) {
	// Create a mock DB and mock objects
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer db.Close()

	repo := NewIngredientRepository(db)

	// Define the ingredient ID and new stock value
	ingredientID := 1
	newStock := 60.0

	// Mock the query for updating the stock
	mock.ExpectExec(`UPDATE ingredients SET current_stock = \$(\d+) WHERE id = \$(\d)`).
		WithArgs(newStock, ingredientID).
		WillReturnResult(sqlmock.NewResult(1, 1)) // Mock success for the update

	// Call the method under test
	err = repo.UpdateStock(context.Background(), nil, ingredientID, newStock)

	// Assertions
	assert.NoError(t, err)

	// Ensure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %v", err)
	}
}

func TestIngredientRepository_CheckLowStockIngredients(t *testing.T) {
	// Create a mock DB and mock objects
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer db.Close()

	repo := NewIngredientRepository(db)

	// Define the expected list of low stock ingredients
	expectedLowStock := []models.Ingredient{
		{ID: 1, Name: "Sugar", TotalStock: 100, CurrentStock: 40},
	}

	// Mock the query for low stock ingredients
	mock.ExpectQuery(`SELECT id, name, total_stock, current_stock FROM ingredients WHERE \(current_stock / total_stock \* 100\) < 50 AND alert_sent = false`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "total_stock", "current_stock"}).
			AddRow(expectedLowStock[0].ID, expectedLowStock[0].Name, expectedLowStock[0].TotalStock, expectedLowStock[0].CurrentStock))

	// Call the method under test
	lowStockIngredients, err := repo.CheckLowStockIngredients(context.Background())

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedLowStock, lowStockIngredients)

	// Ensure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %v", err)
	}
}

func TestIngredientRepository_MarkAlertSent(t *testing.T) {
	// Create a mock DB and mock objects
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer db.Close()

	repo := NewIngredientRepository(db)

	// Define the ingredient ID
	ingredientID := 1

	// Mock the query for marking the alert as sent
	mock.ExpectExec(`UPDATE ingredients SET alert_sent = true WHERE id = \$(\d)`).
		WithArgs(ingredientID).
		WillReturnResult(sqlmock.NewResult(1, 1)) // Mock success for the update

	// Call the method under test
	err = repo.MarkAlertSent(context.Background(), ingredientID)

	// Assertions
	assert.NoError(t, err)

	// Ensure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %v", err)
	}
}
