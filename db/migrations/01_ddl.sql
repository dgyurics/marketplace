CREATE SEQUENCE table_id_seq
    START 1
    INCREMENT 1
    MINVALUE 1
    CACHE 1;

-- Generate 64-bit IDs with parts for timestamp, shard ID, and sequence number.
-- Useful for distributed systems where multiple databases generate IDs
CREATE OR REPLACE FUNCTION gen_id()
RETURNS BIGINT
LANGUAGE plpgsql
AS $$
DECLARE
    our_epoch bigint := 1672531200000; -- custom epoch, 2023-01-01T00:00:00Z in milliseconds
    seq_id bigint; -- sequence ID
    now_millis bigint; -- current time in milliseconds
    shard_id int := 0; -- custom shard ID
    result bigint; -- final ID
BEGIN
    -- Get the next sequence value and modulo by 1024 to get a number between 0 and 1023
    -- Necessary for events where multiple IDs are generated at the same millisecond
    -- FIXME: if more than 1024 IDs are generated in the same millisecond, this will fail
    SELECT nextval('table_id_seq') % 1024 INTO seq_id;

    -- Get the current time in milliseconds
    SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000) INTO now_millis;

    -- Construct the ID
    result := (now_millis - our_epoch) << 23; -- 41 bits for timestamp
    result := result | (shard_id << 10); -- 13 bits for shard ID
    result := result | (seq_id); -- 10 bits for sequence id
    RETURN result; -- FIXME: if most significant bit is 1, this will return a negative number
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

CREATE TABLE IF NOT EXISTS images (
    id BIGINT PRIMARY KEY DEFAULT gen_id(),
    product_id BIGINT NOT NULL,
    image_url TEXT NOT NULL,
    animated BOOLEAN DEFAULT FALSE,
    display_order INT DEFAULT 0,
    alt_text VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

CREATE MATERIALIZED VIEW mv_product AS
SELECT
    p.id,
    p.name,
    p.price,
    p.description,
    COALESCE(
        JSONB_AGG(
            JSONB_BUILD_OBJECT(
                'id', i.id::TEXT,
                'image_url', i.image_url,
                'animated', i.animated,
                'display_order', i.display_order,
                'alt_text', i.alt_text
            )
            ORDER BY i.display_order
        ) FILTER (WHERE i.id IS NOT NULL), '[]'
    ) AS images,
    c.name AS category_name,
    LEAST(inv.quantity, 100) AS quantity
FROM products p
LEFT JOIN images i ON p.id = i.product_id
LEFT JOIN product_categories pc ON p.id = pc.product_id
LEFT JOIN categories c ON pc.category_id = c.id
LEFT JOIN inventory inv ON p.id = inv.product_id
WHERE p.is_deleted = FALSE
GROUP BY p.id, c.name, inv.quantity;

-- Used when fetching a single product by ID
CREATE UNIQUE INDEX idx_mv_product_id
ON mv_product (id);

-- Used when filtering products by category
CREATE INDEX idx_mv_product_category_name
ON mv_product (category_name);

-- Used when sorting products by price
CREATE INDEX idx_mv_product_price
ON mv_product (price);

-- Used when filtering products by category and sorting by price
CREATE INDEX idx_mv_product_category_name_price
ON mv_product (category_name, price);

CREATE TABLE IF NOT EXISTS inventory (
    product_id BIGINT PRIMARY KEY,
    quantity INT NOT NULL DEFAULT 0,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    CHECK (quantity >= 0)
);

-- Used to quickly find products in stock
CREATE INDEX idx_inventory_in_stock
ON inventory (product_id)
WHERE quantity > 0;

CREATE TABLE IF NOT EXISTS product_categories (
    product_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    PRIMARY KEY (product_id, category_id),
    FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE
);

CREATE TYPE user_role_enum AS ENUM ('admin', 'user', 'guest');

CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY DEFAULT gen_id(),
    email VARCHAR(255) UNIQUE,
    password_hash TEXT,
    role user_role_enum DEFAULT 'guest' NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE OR REPLACE VIEW v_users AS
SELECT
    id,
    COALESCE(email, '') AS email,
    COALESCE(password_hash, '') AS password_hash,
    role,
    created_at,
    updated_at
FROM users;


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

CREATE TABLE IF NOT EXISTS password_reset_codes (
    id BIGINT PRIMARY KEY DEFAULT gen_id(),
    user_id BIGINT NOT NULL,
    code_hash TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Used to limit registration to users with an invitation code
CREATE TABLE IF NOT EXISTS invitation_codes (
    code CHAR(6) PRIMARY KEY CHECK (code ~ '^[A-Z0-9]{6}$'),
    used_at TIMESTAMP DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS cart_items (
    user_id BIGINT,
    product_id BIGINT,
    quantity INT NOT NULL,
    unit_price BIGINT NOT NULL,
    PRIMARY KEY (user_id, product_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS addresses (
    id BIGINT PRIMARY KEY DEFAULT gen_id(),
    user_id BIGINT NOT NULL,
    addressee VARCHAR(255),
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    city VARCHAR(255) NOT NULL,
    state_code CHAR(2) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    phone VARCHAR(20),
    is_deleted BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_addresses_is_deleted_false
ON addresses (id)
WHERE is_deleted = FALSE;

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
    address_id BIGINT,
    currency VARCHAR(10) DEFAULT 'usd',
    amount BIGINT NOT NULL DEFAULT 0,
    tax_amount BIGINT NOT NULL DEFAULT 0,
    shipping_amount BIGINT NOT NULL DEFAULT 0,
    total_amount BIGINT NOT NULL DEFAULT 0,
    status order_status_enum NOT NULL DEFAULT 'pending',
    payment_intent_id VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (address_id) REFERENCES addresses(id) ON DELETE RESTRICT
);

-- when checking out, create an order from the user's cart
-- and move the cart items to order_items
CREATE TABLE IF NOT EXISTS order_items (
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    unit_price BIGINT NOT NULL,
    PRIMARY KEY (order_id, product_id),
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE RESTRICT,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT
);

-- View simplifies populating order items array when fetching orders
CREATE OR REPLACE VIEW v_order_items AS
SELECT
    oi.order_id,
    oi.product_id,
    COALESCE(p.description, '') AS description,
    COALESCE(img.image_url, '') AS thumbnail,
    oi.quantity,
    oi.unit_price
FROM order_items oi
JOIN products p ON oi.product_id = p.id
LEFT JOIN images img ON img.product_id = p.id
    AND img.display_order = 0;

CREATE TABLE stripe_events (
    id VARCHAR(255) UNIQUE NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    processed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
REVOKE UPDATE, DELETE ON stripe_events FROM PUBLIC; -- make stripe_events insert only

CREATE OR REPLACE FUNCTION update_or_create_order_from_cart(
    in_user_id BIGINT,
    in_address_id BIGINT
)
RETURNS BIGINT -- Return order ID
LANGUAGE plpgsql
AS $$
DECLARE
    existing_order_id BIGINT;
    cart_rec RECORD;
    existing_item_quantity INT;
BEGIN
    -- 1) Check if user's cart is empty
    IF NOT EXISTS (
        SELECT 1
        FROM cart_items
        WHERE user_id = in_user_id
    ) THEN
        RAISE NOTICE 'User % has an empty cart. No order created or updated.', in_user_id;
        RETURN NULL;
    END IF;

    -- 2) Check if the address exists for the user
    IF NOT EXISTS (
        SELECT 1
        FROM addresses
        WHERE id = in_address_id
        AND user_id = in_user_id
    ) THEN
        RAISE EXCEPTION 'Address % not found for user %', in_address_id, in_user_id;
    END IF;

    -- 3) Check if user has an existing open order
    SELECT id
    INTO existing_order_id
    FROM orders
    WHERE user_id = in_user_id
    AND status = 'pending'
    LIMIT 1;

    -- 4) If an existing order exists, update it
    IF existing_order_id IS NOT NULL THEN
        -- Update the address_id for the existing order
        UPDATE orders
        SET address_id = in_address_id
        WHERE id = existing_order_id;

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
        -- 5) If no existing order, create a new one
        INSERT INTO orders (user_id, status, address_id)
        VALUES (in_user_id, 'pending', in_address_id)
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

    -- 6) Update the order amount
    UPDATE orders
    SET amount = (
            SELECT COALESCE(SUM(quantity * unit_price), 0)
            FROM order_items
            WHERE order_id = existing_order_id
        )
    WHERE id = existing_order_id;

    -- 7) Update the total amount
    UPDATE orders
    SET total_amount = amount
    WHERE id = existing_order_id;

    RETURN existing_order_id;    
EXCEPTION
    WHEN OTHERS THEN
        -- If there's an error, the transaction will be rolled back,
        -- unless you catch and handle it differently here.
        RAISE EXCEPTION 'Error updating or creating order for user %: %', in_user_id, SQLERRM;
END;
$$;
