SET TIME ZONE 'UTC'; -- Important for consistent timestamp handling

-- Table_id_seq is used to generate 10-bit sequence numbers for IDs
-- in the gen_id function. This is necessary to generate unique IDs
-- when multiple IDs are generated at the same millisecond.
CREATE SEQUENCE table_id_seq
    START 1
    INCREMENT 1
    MINVALUE 1
    CACHE 1;

-- Generate 64-bit IDs with parts for timestamp, shard ID, and sequence number.
-- Useful for distributed systems where multiple databases generate     IDs
-- and need to avoid collisions.
CREATE OR REPLACE FUNCTION gen_id()
RETURNS BIGINT
LANGUAGE plpgsql
AS $$
DECLARE
    our_epoch bigint := 1672531200000; -- custom epoch, 2023-01-01T00:00:00Z in milliseconds
    seq_id bigint; -- sequence ID
    now_millis bigint; -- current time in milliseconds
    shard_id int := 0; -- custom shard ID (when using multiple DBs, set this to a unique value per DB)
    result bigint; -- final ID/result
BEGIN
    -- Get the next sequence value and modulo by 1024 to get a number between 0 and 1023
    -- Necessary for events where multiple IDs are generated at the same millisecond
    SELECT nextval('table_id_seq') % 1024 INTO seq_id;

    -- Get the current time in milliseconds
    SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000) INTO now_millis;

    -- Construct the ID
    result := (now_millis - our_epoch) << 23; -- 41 bits for timestamp
    result := result | (shard_id << 10); -- 13 bits for shard ID
    result := result | (seq_id); -- 10 bits for sequence id
    RETURN result;
END;
$$;

CREATE TABLE IF NOT EXISTS categories (
    id BIGINT PRIMARY KEY DEFAULT gen_id(),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS products (
    id BIGINT PRIMARY KEY DEFAULT gen_id(),
    name VARCHAR(255) NOT NULL,
    price BIGINT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TYPE image_type_enum AS ENUM (
    'main',
    'thumbnail',
    'gallery'
);
CREATE TABLE IF NOT EXISTS images (
    id BIGINT PRIMARY KEY DEFAULT gen_id(),
    product_id BIGINT NOT NULL,
    image_url TEXT NOT NULL,
    image_type image_type_enum DEFAULT 'main',
    display_order INT DEFAULT 0,
    alt_text VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS inventory (
    product_id BIGINT PRIMARY KEY,
    quantity INT NOT NULL DEFAULT 0,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    CHECK (quantity >= 0)
);

CREATE TABLE IF NOT EXISTS product_categories (
    product_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    PRIMARY KEY (product_id, category_id),
    FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY DEFAULT gen_id(),
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(255) UNIQUE,
    password_hash TEXT NOT NULL,
    admin BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CHECK (email IS NOT NULL OR phone IS NOT NULL)
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGINT PRIMARY KEY DEFAULT gen_id(),
    user_id BIGINT NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    last_used TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS carts (
    user_id BIGINT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS cart_items (
    user_id BIGINT,
    product_id BIGINT,
    quantity INT NOT NULL,
    unit_price BIGINT NOT NULL,
    PRIMARY KEY (user_id, product_id),
    FOREIGN KEY (user_id) REFERENCES carts(user_id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS addresses (
    id BIGINT PRIMARY KEY DEFAULT gen_id(),
    user_id BIGINT NOT NULL,
    recipient_name VARCHAR(255) NOT NULL,
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    city VARCHAR(255) NOT NULL,
    state_code CHAR(2) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TYPE order_status_enum AS ENUM (
    'created',
    'paid',
    'fulfilled',
    'shipped',
    'delivered',
    'cancelled'
);
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT PRIMARY KEY DEFAULT gen_id(),
    user_id BIGINT,
    shipping_address_id BIGINT,
    total_amount BIGINT NOT NULL, -- total not including tax
    tax_amount BIGINT DEFAULT 0,
    order_status order_status_enum DEFAULT 'created',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (shipping_address_id) REFERENCES addresses(id) ON DELETE SET NULL
);

-- when checking out, create an order from the user's cart
-- and move the cart items to order_items
CREATE TABLE IF NOT EXISTS order_items (
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    unit_price BIGINT NOT NULL,
    PRIMARY KEY (order_id, product_id),
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

CREATE TYPE payment_status_enum AS ENUM (
    'pending',
    'paid',
    'cancelled'
    'refunded'
);
CREATE TABLE payments (
    id BIGINT PRIMARY KEY DEFAULT gen_id(),
    payment_intent_id VARCHAR(255) NOT NULL, -- fixme move this to a separate table. payments table should be vendor agnostic
    client_secret VARCHAR(255) NOT NULL, -- fixme move this to a separate table. payments table should be vendor agnostic  (for frontend confirmation)
    amount INTEGER NOT NULL,
    currency VARCHAR(10) DEFAULT 'usd',
    status payment_status_enum DEFAULT 'pending',
    order_id BIGINT REFERENCES orders(id),
    created_at TIMESTAMP DEFAULT NOW()
);
REVOKE UPDATE, DELETE ON payments FROM PUBLIC; -- make payments insert only

-- Used for Stripe webhook events
CREATE TABLE webhook_events (
    id VARCHAR(255) UNIQUE NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    processed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
REVOKE UPDATE, DELETE ON webhook_events FROM PUBLIC; -- make webhook_events insert only

CREATE OR REPLACE FUNCTION create_order_from_cart(in_user_id BIGINT)
RETURNS BIGINT -- Return the new order's ID
LANGUAGE plpgsql
AS $$
DECLARE
    cart_item_count INT;
    existing_order_id BIGINT;
    new_order_id BIGINT;
    cart_rec RECORD;
BEGIN
    -- 1) Check if user's cart is empty
    SELECT COUNT(*)
    INTO cart_item_count
    FROM cart_items
    WHERE user_id = in_user_id;

    IF cart_item_count = 0 THEN
        RAISE NOTICE 'User % has an empty cart. No order created.', in_user_id;
        RETURN NULL;
    END IF;

    -- 2. Check if user has an existing open order
    SELECT id
    INTO existing_order_id
    FROM orders
    WHERE user_id = in_user_id
    AND order_status = 'created'
    LIMIT 1;

    IF existing_order_id IS NOT NULL THEN
        -- If an order already exists, restore inventory for each item in that order, then delete it
        PERFORM restore_inventory_for_order(existing_order_id);

        DELETE FROM order_items
        WHERE order_id = existing_order_id;

        UPDATE orders
        SET order_status = 'cancelled'
        WHERE id = existing_order_id;
    END IF;

    -- 3) Create a new order (FIXME add shipping address)
    INSERT INTO orders (user_id, total_amount, order_status)
    VALUES (in_user_id, 0, 'created')
    RETURNING id INTO new_order_id;

    -- 4. For each cart item, decrement inventory and insert into order_items
    FOR cart_rec IN
        SELECT product_id, quantity, unit_price
        FROM cart_items
        WHERE user_id = in_user_id
    LOOP
        DECLARE current_inventory INT;
        BEGIN
            SELECT quantity
            INTO current_inventory
            FROM inventory
            WHERE product_id = cart_rec.product_id
            FOR UPDATE; -- lock row

            IF current_inventory IS NULL THEN
                RAISE EXCEPTION 'Product % not found in inventory table', cart_rec.product_id;
            END IF;

            -- If insufficient inventory, throw an error (this will rollback)
            IF current_inventory < cart_rec.quantity THEN
                RAISE EXCEPTION 'Insufficient inventory for product %; needed %, have %',
                    cart_rec.product_id, cart_rec.quantity, current_inventory;
            END IF;

            -- Subtract from inventory
            UPDATE inventory
               SET quantity = quantity - cart_rec.quantity
             WHERE product_id = cart_rec.product_id;

            -- Insert into order_items
            INSERT INTO order_items (order_id, product_id, quantity, unit_price)
            VALUES (new_order_id, cart_rec.product_id, cart_rec.quantity, cart_rec.unit_price);
        END;
    END LOOP;

    -- 5) Update the order total: sum of all order_items
    UPDATE orders
    SET total_amount = (
        SELECT COALESCE(SUM(quantity * unit_price), 0)
        FROM order_items
        WHERE order_id = new_order_id
    )
    WHERE id = new_order_id;

    -- 6) Return newly created order
    RETURN new_order_id;

EXCEPTION
    WHEN OTHERS THEN
        -- If there's an error, the transaction will be rolled back,
        -- unless you catch and handle it differently here.
        RAISE EXCEPTION 'Error creating order for user %: %', in_user_id, SQLERRM;
END;
$$;

CREATE OR REPLACE FUNCTION restore_inventory_for_order(in_order_id BIGINT)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    -- For each item in the order, add back its quantity to inventory
    WITH order_data AS (
        SELECT product_id, quantity
        FROM order_items
        WHERE order_id = in_order_id
    )
    UPDATE inventory i
    SET quantity = i.quantity + order_data.quantity
    FROM order_data
    WHERE i.product_id = order_data.product_id;
END;
$$;

CREATE OR REPLACE FUNCTION mark_order_as_paid(in_order_id BIGINT)
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
    in_user_id BIGINT;
BEGIN
    -- 1) Fetch the user_id associated with the order
    SELECT user_id
    INTO in_user_id
    FROM orders
    WHERE id = in_order_id
      AND order_status = 'created';

    IF NOT FOUND THEN
        RAISE EXCEPTION 'No matching "created" order found for order ID %, or order already finalized.', in_order_id;
    END IF;

    -- 2) Update the order status to 'paid'
    UPDATE orders
    SET order_status = 'paid',
        updated_at = CURRENT_TIMESTAMP
    WHERE id = in_order_id;

    -- 3) Clear the user's cart
    DELETE FROM cart_items
    WHERE user_id = in_user_id;

    DELETE FROM carts
    WHERE user_id = in_user_id;

    RAISE NOTICE 'Order % for user % marked as paid. Cart cleared.', in_order_id, in_user_id;
END;
$$;
