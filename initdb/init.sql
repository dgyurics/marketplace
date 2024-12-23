-- Create the categories table with UUIDs
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL
);

-- Insert categories with hardcoded UUIDs
INSERT INTO categories (id, name, description) VALUES
('3d6f0c4a-75bf-4b9b-9f12-003d6f2f9a1f', 'Category 1', 'This is category 1'),
('81d29ba6-ff4c-4b48-93be-295f31864d5b', 'Category 2', 'This is category 2'),
('4b71dc4e-05e1-4b19-8307-d3dff67dc11f', 'Category 3', 'This is category 3'),
('7ae54a1e-4a4e-40e8-bb0f-c3096d41891f', 'Category 4', 'This is category 4'),
('58c9aaf6-490f-49b6-8c89-64cb7c5e31e3', 'Category 5', 'This is category 5');

-- Create the products table with UUIDs
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    price BIGINT NOT NULL,
    description TEXT NOT NULL
);

CREATE TYPE image_type_enum AS ENUM ('main', 'thumbnail', 'gallery');
CREATE TABLE IF NOT EXISTS product_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    image_url TEXT NOT NULL,
    image_type image_type_enum DEFAULT 'main',
    display_order INT DEFAULT 0,
    alt_text VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS inventory (
    product_id UUID PRIMARY KEY,
    quantity INT NOT NULL DEFAULT 0,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    CHECK (quantity >= 0)
);

-- Insert products with hardcoded UUIDs
INSERT INTO products (id, name, price, description) VALUES
('1c2d6b57-5e1b-4f29-bb38-dbb4b065e5e8', 'Product 1', 1000, 'This is product 1'),
('2a5d7f08-4d2b-4b0a-b8b7-e1bc9f01d898', 'Product 2', 2000, 'This is product 2'),
('3f9b8b1a-9d7c-4b37-8d35-ffa7e00c2a54', 'Product 3', 3000, 'This is product 3'),
('4c6b7d2b-6f0a-4e1d-8a5c-7a4e6f0c9f1b', 'Product 4', 4000, 'This is product 4'),
('5d7a8c1c-7e2c-4f1b-9a6d-8b5f7d0c1a2e', 'Product 5', 5032, 'This is product 5'),
('6e8b9d2d-8f3d-4f2c-a7b8-9c6f8e1d2b3f', 'Product 6', 6000, 'This is product 6'),
('7f9a0e3e-9d4e-4f3d-b8c9-a7e0f9f3c4d5', 'Product 7', 7077, 'This is product 7'),
('8a1b2c4f-ad5f-4b4e-c9d0-b8f1a0f4d6e6', 'Product 8', 8000, 'This is product 8'),
('9b2c3d56-be66-4c5f-da1e-c9e2f0f5e7f7', 'Product 9', 9099, 'This is product 9'),
('af3d4e6a-ce7a-4d6b-eb2f-d0f3a1b6c8a8', 'Product 10', 10000, 'This is product 10');

INSERT INTO inventory (product_id, quantity) VALUES
('1c2d6b57-5e1b-4f29-bb38-dbb4b065e5e8', 100),
('2a5d7f08-4d2b-4b0a-b8b7-e1bc9f01d898', 200),
('3f9b8b1a-9d7c-4b37-8d35-ffa7e00c2a54', 150),
('4c6b7d2b-6f0a-4e1d-8a5c-7a4e6f0c9f1b', 120),
('5d7a8c1c-7e2c-4f1b-9a6d-8b5f7d0c1a2e', 0),
('6e8b9d2d-8f3d-4f2c-a7b8-9c6f8e1d2b3f', 0),
('7f9a0e3e-9d4e-4f3d-b8c9-a7e0f9f3c4d5', 90),
('8a1b2c4f-ad5f-4b4e-c9d0-b8f1a0f4d6e6', 70),
('9b2c3d56-be66-4c5f-da1e-c9e2f0f5e7f7', 110),
('af3d4e6a-ce7a-4d6b-eb2f-d0f3a1b6c8a8', 80);

-- Create the product_categories table with UUIDs as foreign keys
CREATE TABLE IF NOT EXISTS product_categories (
    product_id UUID NOT NULL,
    category_id UUID NOT NULL,
    PRIMARY KEY (product_id, category_id),
    FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE
);

-- Insert data into product_categories using hardcoded UUIDs
INSERT INTO product_categories (product_id, category_id) VALUES 
('1c2d6b57-5e1b-4f29-bb38-dbb4b065e5e8', '3d6f0c4a-75bf-4b9b-9f12-003d6f2f9a1f'),
('2a5d7f08-4d2b-4b0a-b8b7-e1bc9f01d898', '3d6f0c4a-75bf-4b9b-9f12-003d6f2f9a1f'),
('3f9b8b1a-9d7c-4b37-8d35-ffa7e00c2a54', '3d6f0c4a-75bf-4b9b-9f12-003d6f2f9a1f'),
('4c6b7d2b-6f0a-4e1d-8a5c-7a4e6f0c9f1b', '81d29ba6-ff4c-4b48-93be-295f31864d5b'),
('5d7a8c1c-7e2c-4f1b-9a6d-8b5f7d0c1a2e', '81d29ba6-ff4c-4b48-93be-295f31864d5b'),
('6e8b9d2d-8f3d-4f2c-a7b8-9c6f8e1d2b3f', '81d29ba6-ff4c-4b48-93be-295f31864d5b'),
('7f9a0e3e-9d4e-4f3d-b8c9-a7e0f9f3c4d5', '4b71dc4e-05e1-4b19-8307-d3dff67dc11f'),
('8a1b2c4f-ad5f-4b4e-c9d0-b8f1a0f4d6e6', '4b71dc4e-05e1-4b19-8307-d3dff67dc11f'),
('9b2c3d56-be66-4c5f-da1e-c9e2f0f5e7f7', '7ae54a1e-4a4e-40e8-bb0f-c3096d41891f'),
('af3d4e6a-ce7a-4d6b-eb2f-d0f3a1b6c8a8', '58c9aaf6-490f-49b6-8c89-64cb7c5e31e3');

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(255) UNIQUE,
    password_hash TEXT NOT NULL,
    admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CHECK (email IS NOT NULL OR phone IS NOT NULL)
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked BOOLEAN DEFAULT FALSE,
    last_used TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS carts (
    user_id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS cart_items (
    user_id UUID NOT NULL,
    product_id UUID NOT NULL,
    quantity INT NOT NULL,
    unit_price BIGINT NOT NULL,
    PRIMARY KEY (user_id, product_id),
    FOREIGN KEY (user_id) REFERENCES carts(user_id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS shipping_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    recipient_name VARCHAR(255) NOT NULL,
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    city VARCHAR(255) NOT NULL,
    state_code CHAR(2) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
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
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    shipping_address_id UUID,
    total_amount BIGINT NOT NULL, -- total excluding tax
    tax_amount BIGINT DEFAULT 0,
    order_status order_status_enum DEFAULT 'created',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (shipping_address_id) REFERENCES shipping_addresses(id) ON DELETE SET NULL
);

-- store items tied to the order, whether the order is paid or unpaid
CREATE TABLE IF NOT EXISTS order_items (
    order_id UUID NOT NULL,
    product_id UUID NOT NULL,
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
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_intent_id VARCHAR(255) NOT NULL, -- fixme move this to a separate table. payments table should be vendor agnostic
    client_secret VARCHAR(255) NOT NULL, -- fixme move this to a separate table. payments table should be vendor agnostic  (for frontend confirmation)
    amount INTEGER NOT NULL,
    currency VARCHAR(10) DEFAULT 'usd',
    status payment_status_enum DEFAULT 'pending',
    order_id UUID REFERENCES orders(id),
    created_at TIMESTAMP DEFAULT NOW()
);
REVOKE UPDATE, DELETE ON payments FROM PUBLIC; -- make payments insert only

-- Used for Stripe webhook events
CREATE TABLE webhook_events (
    id VARCHAR(255) UNIQUE NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    processed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);
REVOKE UPDATE, DELETE ON webhook_events FROM PUBLIC; -- make webhook_events insert only

CREATE OR REPLACE FUNCTION create_order_from_cart(in_user_id UUID)
RETURNS UUID -- Return the new order's ID
LANGUAGE plpgsql
AS $$
DECLARE
    cart_item_count INT;
    existing_order_id UUID;
    new_order_id UUID;
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

CREATE OR REPLACE FUNCTION restore_inventory_for_order(in_order_id UUID)
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

CREATE OR REPLACE FUNCTION mark_order_as_paid(in_order_id UUID)
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
    in_user_id UUID;
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