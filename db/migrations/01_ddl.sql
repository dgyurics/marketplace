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
    summary text NOT NULL,
    description TEXT,
    details JSONB DEFAULT '{}'::jsonb NOT NULL,
    tax_code VARCHAR(50),
    category_id BIGINT,
    inventory INT NOT NULL DEFAULT 0,
    cart_limit INT,
    is_deleted BOOLEAN DEFAULT FALSE NOT NULL, -- TODO rename to enabled/disabled
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE SET NULL
);

-- Used to quickly find products by category (used by v_products)
CREATE INDEX idx_products_category_is_deleted
ON products (category_id)
WHERE is_deleted = FALSE;

-- Used to quickly find the newest products
CREATE INDEX idx_products_created_at
ON products (created_at DESC)
WHERE is_deleted = FALSE;

CREATE TYPE image_type_enum AS ENUM ('hero', 'thumbnail', 'gallery');

CREATE TABLE IF NOT EXISTS images (
    id BIGINT PRIMARY KEY,
    product_id BIGINT NOT NULL,
    url TEXT NOT NULL,
    type image_type_enum DEFAULT 'hero' NOT NULL,
    source VARCHAR(255) NOT NULL, -- e.g. 115025008992583680.webp
    alt_text VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);
CREATE INDEX idx_images_source ON images(source);

CREATE TYPE user_role_enum AS ENUM ('admin', 'user', 'guest');

CREATE TABLE IF NOT EXISTS pending_users (
    id BIGINT PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    code_hash TEXT NOT NULL,
    used BOOLEAN DEFAULT FALSE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- For quick look-ups by email
CREATE INDEX idx_pending_users_email
ON pending_users(email);

CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    password_hash TEXT,
    role user_role_enum DEFAULT 'guest' NOT NULL,
    requires_setup BOOLEAN,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE OR REPLACE VIEW v_products AS
SELECT
    p.id,
    p.name,
    p.price,
    p.summary,
    COALESCE(p.description, '') AS description,
    p.details,
    p.category_id,
    p.inventory,
    p.cart_limit,
    COALESCE(p.tax_code, '') AS tax_code,
    c.slug AS category_slug,
    COALESCE(imgs.images, '[]') AS images,
    COALESCE(order_stats.total_sold, 0) AS total_sold,
    p.created_at
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN LATERAL (
    SELECT JSONB_AGG(
        JSONB_BUILD_OBJECT(
            'id', i.id::TEXT,
            'url', i.url,
            'type', i.type,
            'updated_at', i.updated_at,
            'alt_text', i.alt_text
        ) ORDER BY i.updated_at DESC
    ) AS images
    FROM images i
    WHERE i.product_id = p.id
) imgs ON TRUE
LEFT JOIN (
    SELECT product_id, sum(oi.quantity) AS total_sold
    FROM order_items oi
    GROUP BY product_id
) order_stats ON order_stats.product_id = p.id
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

CREATE UNLOGGED TABLE IF NOT EXISTS refresh_tokens (
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

CREATE UNLOGGED TABLE IF NOT EXISTS password_reset_codes (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    code_hash TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
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
    country CHAR(2) NOT NULL, -- ISO 3166-1 alpha-2 country code
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
);

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
    address_id BIGINT,
    amount BIGINT NOT NULL DEFAULT 0,
    tax_amount BIGINT NOT NULL DEFAULT 0,
    shipping_amount BIGINT NOT NULL DEFAULT 0,
    total_amount BIGINT NOT NULL DEFAULT 0,
    status order_status_enum NOT NULL DEFAULT 'pending',
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
CREATE INDEX idx_order_items_product_id_quantity
ON order_items(product_id, quantity);

-- View simplifies populating order items array when fetching orders
CREATE OR REPLACE VIEW v_order_items AS
SELECT
    oi.order_id,
    oi.product_id,
    p.name,
    COALESCE(p.summary, '') AS summary,
    COALESCE(p.description, '') AS description,
    COALESCE(i.url, '') AS thumbnail,
    COALESCE(i.alt_text, '') AS alt_text,
    oi.quantity,
    oi.unit_price
FROM order_items oi
JOIN products p ON oi.product_id = p.id
LEFT JOIN images i ON i.product_id = p.id AND i.type = 'thumbnail';

-- required for schedule service
CREATE UNLOGGED TABLE IF NOT EXISTS job_schedules (
    job_name     TEXT PRIMARY KEY,
    last_run_at  TIMESTAMP NOT NULL
);

CREATE UNLOGGED TABLE IF NOT EXISTS rate_limits (
    ip_address INET,
    path TEXT,
    hit_count INTEGER DEFAULT 1,
    expires_at TIMESTAMP,
    PRIMARY KEY (ip_address, path)
);
