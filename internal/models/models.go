package models

import "time"

// Ingredient represents the details of each ingredient.
type Ingredient struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	TotalStock   float64 `json:"total_stock"`
	CurrentStock float64 `json:"current_stock"`
	AlertSent    bool    `json:"alert_sent"`
}

// Product represents the details of each product.
type Product struct {
	ID          int                 `json:"id"`
	Name        string              `json:"name"`
	Ingredients []ProductIngredient `json:"ingredients"`
}

// ProductIngredient represents the relationship between products and ingredients,
// including the amount of ingredient required for each product.
type ProductIngredient struct {
	ProductID    int     `json:"product_id"`
	IngredientID int     `json:"ingredient_id"`
	Amount       float64 `json:"amount"` // Amount of ingredient needed for this product
}

// Order represents a customer order.
type Order struct {
	ID        int         `json:"id"`
	Items     []OrderItem `json:"items"`
	CreatedAt time.Time   `json:"created_at"`
}

// OrderItem represents an individual item in the order.
type OrderItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}
