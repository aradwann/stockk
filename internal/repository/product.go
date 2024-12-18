package repository

import (
	"context"
	"database/sql"
	"fmt"
	"stockk/internal/models"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// GetByID fetches a product by its ID, including its ingredients and amounts
func (r *ProductRepository) GetByID(ctx context.Context, productID int) (*models.Product, error) {
	// Fetch the basic product details
	productQuery := `SELECT id, name FROM products WHERE id = $1`
	var product models.Product
	err := r.db.QueryRowContext(ctx, productQuery, productID).Scan(
		&product.ID,
		&product.Name,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product with ID %d not found", productID)
		}
		return nil, fmt.Errorf("error fetching product: %w", err)
	}

	// Fetch the ingredients for the product
	ingredientsQuery := `
		SELECT pi.product_id, pi.ingredient_id, pi.amount, i.name
		FROM product_ingredients pi
		JOIN ingredients i ON pi.ingredient_id = i.id
		WHERE pi.product_id = $1
	`
	rows, err := r.db.QueryContext(ctx, ingredientsQuery, productID)
	if err != nil {
		return nil, fmt.Errorf("error fetching ingredients for product: %w", err)
	}
	defer rows.Close()

	// Populate the ingredients slice
	for rows.Next() {
		var productIngredient models.ProductIngredient
		if err := rows.Scan(&productIngredient.ProductID, &productIngredient.IngredientID, &productIngredient.Amount); err != nil {
			return nil, fmt.Errorf("error scanning ingredient: %w", err)
		}
		product.Ingredients = append(product.Ingredients, productIngredient)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading ingredient rows: %w", err)
	}

	return &product, nil
}
