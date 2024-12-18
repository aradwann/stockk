-- Remove initial seeded data if needed
DELETE FROM product_ingredients;
DELETE FROM products;
DELETE FROM ingredients;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS product_ingredients;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS ingredients;
