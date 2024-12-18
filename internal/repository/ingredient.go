package repository

import (
	"context"
	"database/sql"
	"fmt"

	"stockk/internal/models"
)

type IngredientRepository struct {
	db *sql.DB
}

func NewIngredientRepository(db *sql.DB) *IngredientRepository {
	return &IngredientRepository{db: db}
}

func (r *IngredientRepository) GetIngredientByID(ctx context.Context, id int) (*models.Ingredient, error) {
	query := `
		SELECT id, name, total_stock, current_stock, alert_sent 
		FROM ingredients 
		WHERE id = $1
	`

	var ingredient models.Ingredient
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ingredient.ID,
		&ingredient.Name,
		&ingredient.TotalStock,
		&ingredient.CurrentStock,
		&ingredient.AlertSent,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ingredient with ID %d not found", id)
		}
		return nil, fmt.Errorf("error fetching ingredient: %w", err)
	}

	return &ingredient, nil
}

func (r *IngredientRepository) GetAllIngredients(ctx context.Context) ([]models.Ingredient, error) {
	query := `
		SELECT id, name, total_stock, current_stock, alert_sent 
		FROM ingredients
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying ingredients: %w", err)
	}
	defer rows.Close()

	var ingredients []models.Ingredient
	for rows.Next() {
		var ingredient models.Ingredient
		if err := rows.Scan(
			&ingredient.ID,
			&ingredient.Name,
			&ingredient.TotalStock,
			&ingredient.CurrentStock,
			&ingredient.AlertSent,
		); err != nil {
			return nil, fmt.Errorf("error scanning ingredient: %w", err)
		}
		ingredients = append(ingredients, ingredient)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading ingredients: %w", err)
	}

	return ingredients, nil
}

func (r *IngredientRepository) UpdateStock(ctx context.Context, ingredientID int, newStock float64) error {
	query := `
		UPDATE ingredients 
		SET current_stock = $1, 
			alert_sent = CASE 
				WHEN (current_stock / total_stock * 100) >= 50 AND ($1 / total_stock * 100) < 50 
				THEN true 
				ELSE alert_sent 
			END
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, newStock, ingredientID)
	if err != nil {
		return fmt.Errorf("error updating ingredient stock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no ingredient found with ID %d", ingredientID)
	}

	return nil
}

func (r *IngredientRepository) CheckLowStockIngredients(ctx context.Context) ([]models.Ingredient, error) {
	query := `
		SELECT id, name, total_stock, current_stock
		FROM ingredients
		WHERE (current_stock / total_stock * 100) < 50 AND alert_sent = false
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying low stock ingredients: %w", err)
	}
	defer rows.Close()

	var lowStockIngredients []models.Ingredient
	for rows.Next() {
		var ingredient models.Ingredient
		if err := rows.Scan(
			&ingredient.ID,
			&ingredient.Name,
			&ingredient.TotalStock,
			&ingredient.CurrentStock,
		); err != nil {
			return nil, fmt.Errorf("error scanning low stock ingredient: %w", err)
		}
		lowStockIngredients = append(lowStockIngredients, ingredient)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading low stock ingredients: %w", err)
	}

	return lowStockIngredients, nil
}

func (r *IngredientRepository) MarkAlertSent(ctx context.Context, ingredientID int) error {
	query := `
		UPDATE ingredients 
		SET alert_sent = true 
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, ingredientID)
	if err != nil {
		return fmt.Errorf("error marking alert as sent: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no ingredient found with ID %d", ingredientID)
	}

	return nil
}
