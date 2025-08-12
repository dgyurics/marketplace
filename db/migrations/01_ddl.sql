CREATE TABLE IF NOT EXISTS categories (
    id BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    parent_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- For tax estimates only
CREATE TABLE IF NOT EXISTS tax_rates (
    country CHAR(2) NOT NULL,
    state VARCHAR(50),
    tax_code VARCHAR(50), -- use NULL for general goods and services tax
    inclusive BOOLEAN DEFAULT FALSE NOT NULL,
    percentage INT NOT NULL, -- scaled by 10000 e.g. 725 = 0.0725
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE (country, state, tax_code)
);

CREATE TABLE IF NOT EXISTS products (
    id BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price BIGINT NOT NULL,
    description TEXT NOT NULL,
    details JSONB DEFAULT '{}'::jsonb NOT NULL,
    tax_code VARCHAR(50),
    category_id BIGINT,
    is_deleted BOOLEAN DEFAULT FALSE NOT NULL, -- TODO rename to enabled/disabled
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE SET NULL
);

-- Used to quickly find products by category (used by v_products)
CREATE INDEX idx_products_category_is_deleted
ON products (category_id, is_deleted)
WHERE is_deleted = FALSE;

CREATE TYPE image_type_enum AS ENUM ('hero', 'thumbnail', 'gallery');

CREATE TABLE IF NOT EXISTS images (
    id BIGINT PRIMARY KEY,
    product_id BIGINT NOT NULL,
    url TEXT NOT NULL,
    type image_type_enum DEFAULT 'hero' NOT NULL,
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

-- Used to quickly find products in stock
CREATE INDEX idx_inventory_in_stock
ON inventory (product_id)
WHERE quantity > 0;

CREATE TYPE user_role_enum AS ENUM ('admin', 'user', 'guest');
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    password_hash TEXT,
    role user_role_enum DEFAULT 'guest' NOT NULL,
    requires_setup BOOLEAN DEFAULT false NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE OR REPLACE VIEW v_products AS
SELECT
    p.id,
    p.name,
    p.price,
    p.description,
    p.details,
    p.category_id,
    COALESCE(p.tax_code, '') AS tax_code,
    c.slug AS category_slug,
    COALESCE(imgs.images, '[]') AS images,
    LEAST(inv.quantity, 100) AS quantity
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN inventory inv ON p.id = inv.product_id
LEFT JOIN LATERAL (
    SELECT JSONB_AGG(
        JSONB_BUILD_OBJECT(
            'id', i.id::TEXT,
            'url', i.url,
            'type', i.type,
            'display_order', i.display_order,
            'alt_text', i.alt_text
        ) ORDER BY i.display_order
    ) AS images
    FROM images i
    WHERE i.product_id = p.id
) imgs ON TRUE
WHERE p.is_deleted = FALSE;

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
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    last_used TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS password_reset_codes (
    id BIGINT PRIMARY KEY,
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
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    addressee VARCHAR(255),
    line1 VARCHAR(255) NOT NULL,
    line2 VARCHAR(255),
    city VARCHAR(255) NOT NULL, -- city, district, suburb, town, village
    state VARCHAR(50) NOT NULL, -- state, county, province, region
    postal_code VARCHAR(20) NOT NULL, -- zip code, postal code
    is_deleted BOOLEAN DEFAULT FALSE NOT NULL, -- when existing order references address, we have to soft delete
    country CHAR(2) NOT NULL, -- ISO 3166-1 alpha-2 country code
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
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
    'canceled'
);
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    email VARCHAR(255),
    address_id BIGINT,
    currency VARCHAR(10) NOT NULL DEFAULT 'usd',
    amount BIGINT NOT NULL DEFAULT 0,
    tax_amount BIGINT NOT NULL DEFAULT 0,
    shipping_amount BIGINT NOT NULL DEFAULT 0,
    total_amount BIGINT NOT NULL DEFAULT 0,
    status order_status_enum NOT NULL DEFAULT 'pending',
    stripe_payment_intent JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (address_id) REFERENCES addresses(id) ON DELETE RESTRICT
);
CREATE INDEX idx_orders_stripe_payment_intent_id
ON orders ((stripe_payment_intent->>'id'));

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
    p.name,
    COALESCE(p.description, '') AS description,
    COALESCE(i.url, '') AS thumbnail,
    COALESCE(i.alt_text, '') AS alt_text,
    oi.quantity,
    oi.unit_price
FROM order_items oi
JOIN products p ON oi.product_id = p.id
LEFT JOIN images i ON i.product_id = p.id AND i.type = 'thumbnail';

CREATE TABLE stripe_events (
    id VARCHAR(255) UNIQUE NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    processed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
REVOKE UPDATE, DELETE ON stripe_events FROM PUBLIC; -- make stripe_events insert only

-- required for schedule service
CREATE TABLE IF NOT EXISTS job_schedules (
    job_name     TEXT PRIMARY KEY,
    last_run_at  TIMESTAMP NOT NULL
);
