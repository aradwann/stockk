-- Remove initial seeded data if needed
DELETE FROM product_ingredients;
DELETE FROM products;
DELETE FROM ingredients;

-- Drop indexes
DROP INDEX IF EXISTS idx_product_ingredients_product_ingredient;
DROP INDEX IF EXISTS idx_ingredients_alert_sent;
DROP INDEX IF EXISTS idx_ingredients_current_stock;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS product_ingredients;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS ingredients;
