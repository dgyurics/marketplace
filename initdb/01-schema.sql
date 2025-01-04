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
    is_deleted BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE INDEX idx_products_is_deleted_false
ON products (id)
WHERE is_deleted = FALSE;

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
    'pending',
    'paid',
    'refunded',
    'fulfilled',
    'shipped',
    'delivered',
    'cancelled'
);
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT PRIMARY KEY DEFAULT gen_id(),
    user_id BIGINT,
    shipping_address_id BIGINT,
    currency VARCHAR(10) DEFAULT 'usd',
    amount BIGINT NOT NULL DEFAULT 0,
    tax_amount BIGINT NOT NULL DEFAULT 0,
    total_amount BIGINT NOT NULL DEFAULT 0,
    status order_status_enum NOT NULL DEFAULT 'pending',
    payment_intent_id VARCHAR(255) NOT NULL DEFAULT '',
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

-- Used for Stripe webhook events
CREATE TABLE webhook_events (
    id VARCHAR(255) UNIQUE NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    processed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
REVOKE UPDATE, DELETE ON webhook_events FROM PUBLIC; -- make webhook_events insert only

CREATE OR REPLACE FUNCTION update_or_create_order_from_cart(in_user_id BIGINT)
RETURNS BIGINT -- Return order ID
LANGUAGE plpgsql
AS $$
DECLARE
    cart_item_count INT;
    existing_order_id BIGINT;
    cart_rec RECORD;
    existing_item_quantity INT;
BEGIN
    -- 1) Check if user's cart is empty
    SELECT COUNT(*)
    INTO cart_item_count
    FROM cart_items
    WHERE user_id = in_user_id;

    IF cart_item_count = 0 THEN
        RAISE NOTICE 'User % has an empty cart. No order created or updated.', in_user_id;
        RETURN NULL;
    END IF;

    -- 2) Check if user has an existing open order
    SELECT id
    INTO existing_order_id
    FROM orders
    WHERE user_id = in_user_id
    AND status = 'pending'
    LIMIT 1;

    -- 3) If an existing order exists, update it
    IF existing_order_id IS NOT NULL THEN
        FOR cart_rec IN
            SELECT product_id, quantity, unit_price
            FROM cart_items
            WHERE user_id = in_user_id
        LOOP
            DECLARE current_inventory INT;
            BEGIN
                -- Check the existing reserved quantity in the order
                SELECT quantity
                INTO existing_item_quantity
                FROM order_items
                WHERE order_id = existing_order_id
                AND product_id = cart_rec.product_id;

                -- If no matching item in the order, treat reserved quantity as 0
                IF NOT FOUND THEN
                    existing_item_quantity := 0;
                END IF;

                -- Check inventory, accounting for already reserved items
                SELECT quantity
                INTO current_inventory
                FROM inventory
                WHERE product_id = cart_rec.product_id
                FOR UPDATE; -- lock row

                IF current_inventory IS NULL THEN
                    RAISE EXCEPTION 'Product % not found in inventory table', cart_rec.product_id;
                END IF;

                -- Effective inventory available is current minus already reserved
                IF current_inventory + existing_item_quantity < cart_rec.quantity THEN
                    RAISE EXCEPTION 'Insufficient inventory for product %; needed %, available %',
                        cart_rec.product_id, cart_rec.quantity, current_inventory + existing_item_quantity;
                END IF;

                -- Update inventory only for new quantities added
                UPDATE inventory
                SET quantity = quantity - (cart_rec.quantity - existing_item_quantity)
                WHERE product_id = cart_rec.product_id;

                -- Insert or update order_items for the existing order
                INSERT INTO order_items (order_id, product_id, quantity, unit_price)
                VALUES (existing_order_id, cart_rec.product_id, cart_rec.quantity, cart_rec.unit_price)
                ON CONFLICT (order_id, product_id) DO UPDATE
                SET quantity = EXCLUDED.quantity, unit_price = EXCLUDED.unit_price;
            END;
        END LOOP;
    ELSE
        -- 4) If no existing order, create a new one
        INSERT INTO orders (user_id, status)
        VALUES (in_user_id, 'pending')
        RETURNING id INTO existing_order_id;

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
                VALUES (existing_order_id, cart_rec.product_id, cart_rec.quantity, cart_rec.unit_price);
            END;
        END LOOP;
    END IF;

    -- Update the order amount
    UPDATE orders
    SET amount = (
            SELECT COALESCE(SUM(quantity * unit_price), 0)
            FROM order_items
            WHERE order_id = existing_order_id
        )
    WHERE id = existing_order_id;

    -- Update the total amount
    UPDATE orders
    SET total_amount = amount + tax_amount
    WHERE id = existing_order_id;

    RETURN existing_order_id;    
EXCEPTION
    WHEN OTHERS THEN
        -- If there's an error, the transaction will be rolled back,
        -- unless you catch and handle it differently here.
        RAISE EXCEPTION 'Error updating or creating order for user %: %', in_user_id, SQLERRM;
END;
$$;
