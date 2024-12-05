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
    price NUMERIC(10, 2) NOT NULL,
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
('1c2d6b57-5e1b-4f29-bb38-dbb4b065e5e8', 'Product 1', 10.00, 'This is product 1'),
('2a5d7f08-4d2b-4b0a-b8b7-e1bc9f01d898', 'Product 2', 20.00, 'This is product 2'),
('3f9b8b1a-9d7c-4b37-8d35-ffa7e00c2a54', 'Product 3', 30.00, 'This is product 3'),
('4c6b7d2b-6f0a-4e1d-8a5c-7a4e6f0c9f1b', 'Product 4', 40.00, 'This is product 4'),
('5d7a8c1c-7e2c-4f1b-9a6d-8b5f7d0c1a2e', 'Product 5', 50.32, 'This is product 5'),
('6e8b9d2d-8f3d-4f2c-a7b8-9c6f8e1d2b3f', 'Product 6', 60.00, 'This is product 6'),
('7f9a0e3e-9d4e-4f3d-b8c9-a7e0f9f3c4d5', 'Product 7', 70.77, 'This is product 7'),
('8a1b2c4f-ad5f-4b4e-c9d0-b8f1a0f4d6e6', 'Product 8', 80.00, 'This is product 8'),
('9b2c3d56-be66-4c5f-da1e-c9e2f0f5e7f7', 'Product 9', 90.99, 'This is product 9'),
('af3d4e6a-ce7a-4d6b-eb2f-d0f3a1b6c8a8', 'Product 10', 100.00, 'This is product 10');

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

CREATE TABLE IF NOT EXISTS inventory_reservations (
    product_id UUID NOT NULL,
    user_id UUID NOT NULL,
    reserved_quantity INT NOT NULL,
    reservation_expiration TIMESTAMP DEFAULT (CURRENT_TIMESTAMP + INTERVAL '15 minutes'),
    FOREIGN KEY (product_id) REFERENCES inventory(product_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, user_id)
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
    unit_price NUMERIC(10, 2) NOT NULL,
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

CREATE TYPE order_status_enum AS ENUM ('created', 'paid', 'fulfilled', 'cancelled');
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    shipping_address_id UUID,
    total_amount NUMERIC(10, 2) NOT NULL,
    tax_amount NUMERIC(10, 2) DEFAULT 0,
    order_status order_status_enum DEFAULT 'created',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (shipping_address_id) REFERENCES shipping_addresses(id) ON DELETE SET NULL
);

CREATE TYPE payment_status_enum AS ENUM (
    'requires_payment_method',
    'requires_confirmation',
    'requires_action',
    'processing',
    'succeeded',
    'canceled',
    'unknown' -- Fallback for unrecognized statuses
);
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_intent_id VARCHAR(255) NOT NULL,
    client_secret VARCHAR(255) NOT NULL, -- for frontend confirmation
    amount INTEGER NOT NULL,
    currency VARCHAR(10) DEFAULT 'usd',
    status payment_status_enum DEFAULT 'requires_payment_method',
    order_id UUID REFERENCES orders(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION reserve_cart_items(usrid UUID)
RETURNS TEXT AS $$
DECLARE
    cart_item_count INT;
    updated_row_count INT;
BEGIN
    -- Check if there are items in the user's cart
    SELECT COUNT(*)
    INTO cart_item_count
    FROM cart_items
    WHERE user_id = usrid;

    -- Return 'empty_cart' if no items in the cart
    IF cart_item_count = 0 THEN
        RETURN 'empty_cart';
    END IF;

    -- Restore inventory from existing reservations, if any
    WITH restored_inventory AS (
        UPDATE inventory
        SET quantity = inventory.quantity + inventory_reservations.reserved_quantity
        FROM inventory_reservations
        WHERE inventory.product_id = inventory_reservations.product_id
          AND inventory_reservations.user_id = usrid
        RETURNING inventory_reservations.product_id
    )
    -- Delete the old reservations after restoring inventory
    DELETE FROM inventory_reservations WHERE user_id = usrid;

    -- Update inventory and attempt to reserve items
    WITH updated_inventory AS (
        UPDATE inventory
        SET quantity = inventory.quantity - cart_items.quantity
        FROM cart_items
        WHERE inventory.product_id = cart_items.product_id
          AND cart_items.user_id = usrid
          AND inventory.quantity >= cart_items.quantity
        RETURNING cart_items.product_id, cart_items.quantity
    )
    INSERT INTO inventory_reservations (product_id, user_id, reserved_quantity, reservation_expiration)
    SELECT updated_inventory.product_id, usrid, updated_inventory.quantity, CURRENT_TIMESTAMP + INTERVAL '15 minutes'
    FROM updated_inventory
    ON CONFLICT (product_id, user_id)
    DO UPDATE SET reserved_quantity = EXCLUDED.reserved_quantity,
                  reservation_expiration = EXCLUDED.reservation_expiration;

    -- Get the number of rows updated
    GET DIAGNOSTICS updated_row_count = ROW_COUNT;

    -- If no rows were updated, return 'insufficient_inventory'
    IF updated_row_count = 0 THEN
        RETURN 'insufficient_inventory';
    END IF;

    -- If the updated rows don't match the cart item count, rollback
    IF updated_row_count <> cart_item_count THEN
        RAISE EXCEPTION 'Insufficient inventory for some items in cart. Reservation failed.';
    END IF;

    -- Return 'success' if everything was reserved successfully
    RETURN 'success';

EXCEPTION
    WHEN OTHERS THEN
        -- Catch any unexpected error and rollback
        RAISE EXCEPTION 'An error occurred during reservation: %', SQLERRM;
END;
$$ LANGUAGE plpgsql;
