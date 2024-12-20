package repository

import (
	"context"
	"database/sql"
	"log/slog"

	"stockk/internal/errors"
	"stockk/internal/models"
)

type IngredientRepository interface {
	GetIngredientByID(ctx context.Context, tx Transaction, ingredientID int) (*models.Ingredient, error)
	UpdateStock(ctx context.Context, tx Transaction, ingredientID int, newStock float64) error
	CheckLowStockIngredients(ctx context.Context) ([]models.Ingredient, error)
	MarkAlertSent(ctx context.Context, ingredientID int) error
}

type ingredientRepository struct {
	db *sql.DB
}

func NewIngredientRepository(db *sql.DB) IngredientRepository {
	return &ingredientRepository{db: db}
}

var _ IngredientRepository = (*ingredientRepository)(nil)

func (r *ingredientRepository) GetIngredientByID(ctx context.Context, tx Transaction, ingredientID int) (*models.Ingredient, error) {
	query := `
		SELECT id, name, total_stock, current_stock, alert_sent 
		FROM ingredients 
		WHERE id = $1
	`

	var ingredient models.Ingredient
	var err error

	// Use transaction if provided
	if tx != nil {
		err = tx.QueryRowContext(ctx, query, ingredientID).Scan(
			&ingredient.ID,
			&ingredient.Name,
			&ingredient.TotalStock,
			&ingredient.CurrentStock,
			&ingredient.AlertSent,
		)
	} else {
		err = r.db.QueryRowContext(ctx, query, ingredientID).Scan(
			&ingredient.ID,
			&ingredient.Name,
			&ingredient.TotalStock,
			&ingredient.CurrentStock,
			&ingredient.AlertSent,
		)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Wrap(errors.ErrNotFound, "ingredient not found")
		}
		slog.Error("failed to retrieve ingredient", "ingredientID", ingredientID, "error", err)
		return nil, errors.Wrap(errors.ErrInternalServer, "query failed")
	}

	return &ingredient, nil
}

func (r *ingredientRepository) UpdateStock(ctx context.Context, tx Transaction, ingredientID int, newStock float64) error {
	query := `
		UPDATE ingredients 
		SET current_stock = $1 
		WHERE id = $2
	`
	var err error
	var result sql.Result
	if tx != nil {
		result, err = tx.ExecContext(ctx, query, newStock, ingredientID)
	} else {
		result, err = r.db.ExecContext(ctx, query, newStock, ingredientID)
	}
	if err != nil {
		slog.Error("failed to update ingredient stock", "ingredientID", ingredientID, "error", err)
		return errors.Wrap(errors.ErrInternalServer, "query failed")

	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("failed to update ingredient stock", "ingredientID", ingredientID, "error", err)
		return errors.Wrap(errors.ErrInternalServer, "query failed")
	}

	if rowsAffected == 0 {
		return errors.Wrap(errors.ErrNotFound, "ingredient not found")
	}

	return nil
}

func (r *ingredientRepository) CheckLowStockIngredients(ctx context.Context) ([]models.Ingredient, error) {
	query := `
		SELECT id, name, total_stock, current_stock
		FROM ingredients
		WHERE (current_stock / total_stock * 100) < 50 AND alert_sent = false
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		slog.Error("failed to retrieve low stock ingredients", "error", err)
		return nil, errors.Wrap(errors.ErrInternalServer, "query failed")
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
			slog.Error("failed to retrieve low stock ingredients", "error", err)
			return nil, errors.Wrap(errors.ErrInternalServer, "query failed")
		}
		lowStockIngredients = append(lowStockIngredients, ingredient)
	}

	if err = rows.Err(); err != nil {
		slog.Error("failed to retrieve low stock ingredients", "error", err)
		return nil, errors.Wrap(errors.ErrInternalServer, "query failed")
	}

	return lowStockIngredients, nil
}

func (r *ingredientRepository) MarkAlertSent(ctx context.Context, ingredientID int) error {
	query := `
		UPDATE ingredients 
		SET alert_sent = true 
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, ingredientID)
	if err != nil {
		slog.Error("failed to update ingredient alert status", "error", err)
		return errors.Wrap(errors.ErrInternalServer, "query failed")

	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("failed to update ingredient alert status", "error", err)
		return errors.Wrap(errors.ErrInternalServer, "query failed")

	}

	if rowsAffected == 0 {
		return errors.Wrap(errors.ErrNotFound, "ingredient not found")
	}

	return nil
}
