package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"stockk/internal/errors"
	internalErrors "stockk/internal/errors"
	"stockk/internal/models"
)

type ProductRepository interface {
	GetProductById(ctx context.Context, tx Transaction, productId int) (*models.Product, error)
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

var _ ProductRepository = (*productRepository)(nil)

// GetProductById fetches a product by its ID, including its ingredients and amounts
func (r *productRepository) GetProductById(ctx context.Context, tx Transaction, productID int) (*models.Product, error) {
	// Fetch the basic product details
	productQuery := `SELECT id, name FROM products WHERE id = $1`
	var product models.Product
	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, productQuery, productID).Scan(
			&product.ID,
			&product.Name,
		)
	} else {
		err = r.db.QueryRowContext(ctx, productQuery, productID).Scan(
			&product.ID,
			&product.Name,
		)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, internalErrors.NewAppError(internalErrors.ErrCodeNotFound, "Resource not found", fmt.Sprintf("Product with ID %d not found", productID))
		}
		slog.Error("failed to retrieve product", "error", err)
		return nil, errors.Wrap(errors.ErrInternalServer, "query failed")
	}

	// Fetch the ingredients for the product
	ingredientsQuery := `
		SELECT pi.product_id, pi.ingredient_id, pi.amount
		FROM product_ingredients pi
		JOIN ingredients i ON pi.ingredient_id = i.id
		WHERE pi.product_id = $1
	`
	var rows *sql.Rows
	if tx != nil {
		rows, err = tx.QueryContext(ctx, ingredientsQuery, productID)

	} else {
		rows, err = r.db.QueryContext(ctx, ingredientsQuery, productID)

	}
	if err != nil {
		slog.Error("failed to retrieve order ingredients", "error", err)
		return nil, errors.Wrap(errors.ErrInternalServer, "query failed")
	}
	defer rows.Close()

	// Populate the ingredients slice
	for rows.Next() {
		var productIngredient models.ProductIngredient
		if err := rows.Scan(&productIngredient.ProductID, &productIngredient.IngredientID, &productIngredient.Amount); err != nil {
			slog.Error("failed to retrieve order ingredients", "error", err)
			return nil, errors.Wrap(errors.ErrInternalServer, "query failed")
		}
		product.Ingredients = append(product.Ingredients, productIngredient)
	}

	if err := rows.Err(); err != nil {
		slog.Error("failed to retrieve order ingredients", "error", err)
		return nil, errors.Wrap(errors.ErrInternalServer, "query failed")
	}

	return &product, nil
}
