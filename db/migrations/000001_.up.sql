CREATE TABLE ingredients (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    total_stock NUMERIC(10, 2) NOT NULL CHECK (total_stock >= 0),
    current_stock NUMERIC(10, 2) NOT NULL CHECK (current_stock >= 0),
    alert_sent BOOLEAN DEFAULT FALSE
);

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE TABLE product_ingredients (
    product_id INTEGER REFERENCES products(id),
    ingredient_id INTEGER REFERENCES ingredients(id),
    amount NUMERIC(10, 2) NOT NULL CHECK (amount >= 0),
    PRIMARY KEY (product_id, ingredient_id)
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE order_items (
    order_id INTEGER REFERENCES orders(id),
    product_id INTEGER REFERENCES products(id),
    quantity INTEGER NOT NULL,
    PRIMARY KEY (order_id, product_id)
);

-- Seed initial data
INSERT INTO ingredients (name, total_stock, current_stock) VALUES 
('Beef', 20000, 20000),
('Cheese', 5000, 5000),
('Onion', 1000, 1000);

INSERT INTO products (name) VALUES 
('Burger');

INSERT INTO product_ingredients (product_id, ingredient_id, amount) VALUES
(1, 1, 150),  -- 150g Beef
(1, 2, 30),   -- 30g Cheese
(1, 3, 20);   -- 20g Onion