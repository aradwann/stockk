CREATE OR REPLACE FUNCTION create_order_with_items(
    product_ids INTEGER[],
    quantities INTEGER[]
)
RETURNS VOID AS $$
DECLARE
    order_id INTEGER;
    i INTEGER;
    quantity_variable INTEGER;
    ingredient_id_variable INTEGER; -- Alias for ingredient_id from SELECT
    product_id_variable INTEGER;    -- Alias for product_id from SELECT
    amount_per_unit NUMERIC(10, 2);
    total_amount_needed NUMERIC(10, 2);
    current_stock_variable NUMERIC(10, 2);
BEGIN
    -- Validate input array lengths
    IF array_length(product_ids, 1) IS DISTINCT FROM array_length(quantities, 1) THEN
        RAISE EXCEPTION 'Product IDs and quantities must have the same length';
    END IF;

    -- Create the order
    INSERT INTO orders DEFAULT VALUES RETURNING id INTO order_id;

    -- Loop through the array
    FOR i IN 1..array_length(product_ids, 1) LOOP
        product_id_variable := product_ids[i];
        quantity_variable := quantities[i];

        -- Check ingredient stock and deduct
        FOR ingredient_id_variable, amount_per_unit IN
            SELECT pi.ingredient_id, pi.amount
            FROM product_ingredients pi
            WHERE pi.product_id = product_id_variable
        LOOP
            total_amount_needed := amount_per_unit * quantity_variable;

            SELECT current_stock
            INTO current_stock_variable
            FROM ingredients
            WHERE id = ingredient_id_variable;

            IF current_stock_variable < total_amount_needed THEN
                RAISE EXCEPTION 'Insufficient stock for ingredient %: needed %, available %', ingredient_id_variable, total_amount_needed, current_stock_variable;
            END IF;

            UPDATE ingredients
            SET current_stock = current_stock_variable - total_amount_needed
            WHERE id = ingredient_id_variable;
        END LOOP;

        -- Insert into order_items
        INSERT INTO order_items (order_id, product_id, quantity)
        VALUES (order_id, product_id_variable, quantity_variable);
    END LOOP;
END;
$$ LANGUAGE plpgsql;
