CREATE TABLE categories (
    id BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    parent_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- For tax estimates only
CREATE TABLE tax_rates (
    country CHAR(2) NOT NULL,
    state VARCHAR(10),
    tax_code VARCHAR(50), -- use NULL for general goods and services tax
    inclusive BOOLEAN DEFAULT FALSE NOT NULL,
    percentage INT NOT NULL, -- scaled by 10000 e.g. 725 = 0.0725
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE (country, state, tax_code)
);

CREATE TYPE product_status_enum AS ENUM ('draft', 'published', 'archived');

CREATE TABLE products (
    id BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price BIGINT NOT NULL,
    summary TEXT,
    description TEXT,
    details JSONB DEFAULT '{}'::jsonb NOT NULL,
    tax_code VARCHAR(50),
    category_id BIGINT,
    inventory INT NOT NULL DEFAULT 0,
    cart_limit INT,
    negotiable BOOLEAN DEFAULT FALSE NOT NULL, -- Defines whether the product accepts offers
    featured BOOLEAN DEFAULT FALSE NOT NULL, -- Defines whether the product should be highlighted in the UI homepage
    sort_order INT DEFAULT 0 NOT NULL, -- Higher numbers appear first
    status product_status_enum NOT NULL DEFAULT 'published',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE SET NULL
);

-- Used to quickly find products by category (used by v_products)
CREATE INDEX idx_products_category_status
ON products (category_id)
WHERE status <> 'archived';

-- Used to quickly find the newest products
CREATE INDEX idx_products_created_at
ON products (created_at DESC)
WHERE status <> 'archived';

CREATE TYPE image_type_enum AS ENUM ('hero', 'thumbnail', 'gallery');

CREATE TABLE images (
    id BIGINT PRIMARY KEY,
    product_id BIGINT NOT NULL,
    url TEXT NOT NULL,
    type image_type_enum DEFAULT 'hero' NOT NULL,
    source VARCHAR(255) NOT NULL, -- e.g. 115025008992583680.webp
    alt_text VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
);
CREATE INDEX idx_images_source ON images (source);
CREATE INDEX idx_images_product_id ON images(product_id);

CREATE TYPE user_role_enum AS ENUM ('admin', 'user', 'guest', 'member', 'staff', 'system');
CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    password_hash TEXT,
    role user_role_enum NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE INDEX idx_users_guest_created_at
ON users (created_at)
WHERE role = 'guest';

CREATE TABLE registration_codes (
  code CHAR(6) PRIMARY KEY,
  user_id BIGINT NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
-- For cleanup queries
CREATE INDEX idx_registration_codes_expires ON registration_codes (expires_at);
CREATE INDEX idx_registration_codes_user_id
ON registration_codes (user_id);

CREATE OR REPLACE VIEW v_users AS
SELECT
    id,
    COALESCE(email, '') AS email,
    COALESCE(password_hash, '') AS password_hash,
    role,
    created_at,
    updated_at
FROM users;

CREATE UNLOGGED TABLE refresh_tokens (
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
CREATE INDEX idx_refresh_tokens_user_id_last_used
ON refresh_tokens (user_id, last_used);

CREATE UNLOGGED TABLE password_reset_codes (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    code_hash TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
CREATE INDEX idx_password_reset_codes_user_id
ON password_reset_codes (user_id);

CREATE TABLE cart_items (
    user_id BIGINT,
    product_id BIGINT,
    quantity INT NOT NULL,
    unit_price BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, product_id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
);

CREATE TABLE addresses (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(255),
    line1 VARCHAR(255) NOT NULL,
    line2 VARCHAR(255),
    city VARCHAR(255) NOT NULL, -- city, district, suburb, town, village
    state VARCHAR(10), -- state, county, province, region
    postal_code VARCHAR(20) NOT NULL,
    country CHAR(2) NOT NULL, -- ISO 3166-1 alpha-2 country code
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
CREATE INDEX idx_addresses_user_id ON addresses(user_id);

CREATE TYPE order_status_enum AS ENUM (
    'draft',
    'pending',
    'paid',
    'shipped',
    'delivered',
    'canceled',
    'refunded',
    'requested'
);
CREATE TYPE payment_method_enum AS ENUM ('stripe', 'on_delivery');
CREATE TYPE payment_status_enum AS ENUM ('unpaid', 'paid', 'refunded');
CREATE TABLE orders (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    address_id BIGINT NOT NULL,
    amount BIGINT NOT NULL DEFAULT 0,
    tax_amount BIGINT NOT NULL DEFAULT 0,
    shipping_amount BIGINT NOT NULL DEFAULT 0,
    total_amount BIGINT NOT NULL DEFAULT 0,
    status order_status_enum NOT NULL DEFAULT 'draft',
    idempotency_key VARCHAR(64),
    payment_method payment_method_enum, -- NULL when status is 'draft'
    payment_status payment_status_enum NOT NULL DEFAULT 'unpaid',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (address_id) REFERENCES addresses(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX idx_orders_user_id_unique
ON orders (user_id)
WHERE status = 'draft' OR (status = 'pending' AND payment_method = 'stripe');
CREATE UNIQUE INDEX idx_orders_idempotency_key
ON orders (idempotency_key)
WHERE idempotency_key IS NOT NULL;
CREATE INDEX idx_orders_user_id
ON orders (user_id);
CREATE INDEX idx_orders_address_id
ON orders (address_id);

CREATE TABLE order_items (
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    unit_price BIGINT NOT NULL,
    PRIMARY KEY (order_id, product_id),
    FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE RESTRICT
);
CREATE INDEX idx_order_items_product_id_quantity
ON order_items (product_id, quantity);

CREATE TABLE requested_items (
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    unit_price BIGINT NOT NULL,
    PRIMARY KEY (order_id, product_id),
    FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE RESTRICT
);

-- View simplifies fetching product details including total sold count
-- Product images are loaded directly from images by the repositories.
CREATE OR REPLACE VIEW v_products AS
SELECT
    p.id,
    p.name,
    p.price,
    COALESCE(p.summary, '') AS summary,
    COALESCE(p.description, '') AS description,
    p.details,
    p.category_id,
    p.inventory,
    p.cart_limit,
    COALESCE(p.tax_code, '') AS tax_code,
    p.featured,
    p.negotiable,
    p.sort_order,
    c.slug AS category_slug,
    COALESCE(order_stats.total_sold, 0) AS total_sold,
    p.created_at,
    p.status
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN (
    SELECT product_id, sum(oi.quantity) AS total_sold
    FROM order_items oi
    GROUP BY product_id
) order_stats ON order_stats.product_id = p.id
WHERE p.status <> 'archived';

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
CREATE UNLOGGED TABLE job_schedules (
    job_name     TEXT PRIMARY KEY,
    last_run_at  TIMESTAMP NOT NULL
);

CREATE UNLOGGED TABLE rate_limits (
    ip_address INET,
    path TEXT,
    hit_count INTEGER DEFAULT 1,
    expires_at TIMESTAMP,
    PRIMARY KEY (ip_address, path)
);

-- Defines shipping zones supported by the marketplace
CREATE TABLE shipping_zones (
    id BIGINT PRIMARY KEY,
    country CHAR(2) NOT NULL,
    state VARCHAR(10) NOT NULL DEFAULT '',  -- empty string to allow UNIQUE constraint to work
    postal_code VARCHAR(20) NOT NULL DEFAULT '', -- empty string to allow UNIQUE constraint to work
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE(country, state, postal_code)
);
CREATE UNIQUE INDEX shipping_zones_postal_code_unique
ON shipping_zones (postal_code)
WHERE postal_code != '';

-- Defines shipping zones NOT supported by the marketplace
CREATE TABLE shipping_exclusions (
    id BIGINT PRIMARY KEY,
    country CHAR(2) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE(country, postal_code)
);

CREATE TYPE offer_status_enum AS ENUM (
    'pending',
    'accepted',
    'rejected',
    'canceled',
    'completed'
);
CREATE TABLE offers (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  product_id BIGINT NOT NULL,
  amount BIGINT NOT NULL,
  status offer_status_enum NOT NULL,
  comment TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE RESTRICT
);
CREATE INDEX idx_offers_user_id
ON offers (user_id);

CREATE TYPE conversation_type_enum AS ENUM ('support', 'notification');

CREATE TABLE conversations (
    id BIGINT PRIMARY KEY,
    type conversation_type_enum NOT NULL,
    subject TEXT NOT NULL,
    recipient_id BIGINT NOT NULL,
    recipient_last_read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (recipient_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_conversations_recipient_id
ON conversations (recipient_id);

CREATE TABLE messages (
    id BIGINT PRIMARY KEY,
    sender_id BIGINT NOT NULL,
    conversation_id BIGINT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);
CREATE INDEX idx_messages_conversation_id ON messages(conversation_id, created_at DESC);
CREATE INDEX idx_messages_sender_id
ON messages (sender_id);

-- Accounts
INSERT INTO users (id, email, password_hash, role, verified) VALUES (114717860949983232, 'admin@marketplace.com', '$2a$10$foTM0ULLDB3KW1uhsCkTkesYQknpocAbMNT4Xfdp.vQ1C1U4HLfai', 'admin', true);
INSERT INTO users (id, role, verified) VALUES (1,'system',true);

-- Tax rates
-- United States
INSERT INTO tax_rates (country, state, inclusive, percentage)
VALUES
('US', 'AL', FALSE, 921),  -- Alabama
('US', 'AK', FALSE, 176),  -- Alaska (no state tax, local only)
('US', 'AZ', FALSE, 838),  -- Arizona
('US', 'AR', FALSE, 944),  -- Arkansas
('US', 'CA', FALSE, 882),  -- California
('US', 'CO', FALSE, 777),  -- Colorado
('US', 'CT', FALSE, 635),  -- Connecticut
('US', 'DE', FALSE, 0),    -- Delaware
('US', 'FL', FALSE, 702),  -- Florida
('US', 'GA', FALSE, 737),  -- Georgia
('US', 'HI', FALSE, 444),  -- Hawaii (GET, not traditional sales tax)
('US', 'ID', FALSE, 606),  -- Idaho
('US', 'IL', FALSE, 882),  -- Illinois
('US', 'IN', FALSE, 700),  -- Indiana
('US', 'IA', FALSE, 688),  -- Iowa
('US', 'KS', FALSE, 869),  -- Kansas
('US', 'KY', FALSE, 600),  -- Kentucky
('US', 'LA', FALSE, 988),  -- Louisiana
('US', 'ME', FALSE, 555),  -- Maine
('US', 'MD', FALSE, 600),  -- Maryland
('US', 'MA', FALSE, 625),  -- Massachusetts
('US', 'MI', FALSE, 600),  -- Michigan
('US', 'MN', FALSE, 741),  -- Minnesota
('US', 'MS', FALSE, 707),  -- Mississippi
('US', 'MO', FALSE, 820),  -- Missouri
('US', 'MT', FALSE, 0),    -- Montana
('US', 'NE', FALSE, 691),  -- Nebraska
('US', 'NV', FALSE, 836),  -- Nevada
('US', 'NH', FALSE, 0),    -- New Hampshire
('US', 'NJ', FALSE, 666),  -- New Jersey
('US', 'NM', FALSE, 785),  -- New Mexico
('US', 'NY', FALSE, 852),  -- New York
('US', 'NC', FALSE, 698),  -- North Carolina
('US', 'ND', FALSE, 698),  -- North Dakota
('US', 'OH', FALSE, 724),  -- Ohio
('US', 'OK', FALSE, 947),  -- Oklahoma
('US', 'OR', FALSE, 0),    -- Oregon (no sales tax)
('US', 'PA', FALSE, 634),  -- Pennsylvania
('US', 'RI', FALSE, 700),  -- Rhode Island
('US', 'SC', FALSE, 747),  -- South Carolina
('US', 'SD', FALSE, 642),  -- South Dakota
('US', 'TN', FALSE, 955),  -- Tennessee
('US', 'TX', FALSE, 820),  -- Texas
('US', 'UT', FALSE, 712),  -- Utah
('US', 'VT', FALSE, 620),  -- Vermont
('US', 'VA', FALSE, 581),  -- Virginia
('US', 'WA', FALSE, 940),  -- Washington
('US', 'WV', FALSE, 654),  -- West Virginia
('US', 'WI', FALSE, 543),  -- Wisconsin
('US', 'WY', FALSE, 539),  -- Wyoming
('US', 'DC', FALSE, 620),  -- District of Columbia

-- U.S. Territories
('US', 'AS', FALSE, 0),    -- American Samoa
('US', 'GU', FALSE, 400),  -- Guam
('US', 'MP', FALSE, 0),    -- Northern Mariana Islands
('US', 'PR', FALSE, 1050), -- Puerto Rico
('US', 'VI', FALSE, 600);  -- U.S. Virgin Islands

-- Canada
INSERT INTO tax_rates (country, state, inclusive, percentage)
VALUES
('CA', 'AB', TRUE, 500),   -- Alberta
('CA', 'BC', TRUE, 1200),  -- British Columbia
('CA', 'MB', TRUE, 1200),  -- Manitoba
('CA', 'NB', TRUE, 1500),  -- New Brunswick
('CA', 'NL', TRUE, 1500),  -- Newfoundland and Labrador
('CA', 'NS', TRUE, 1500),  -- Nova Scotia
('CA', 'NT', TRUE, 500),   -- Northwest Territories
('CA', 'NU', TRUE, 500),   -- Nunavut
('CA', 'ON', TRUE, 1300),  -- Ontario
('CA', 'PE', TRUE, 1500),  -- Prince Edward Island
('CA', 'QC', TRUE, 1498),  -- Quebec
('CA', 'SK', TRUE, 1100),  -- Saskatchewan
('CA', 'YT', TRUE, 500);   -- Yukon (GST only)

-- Scandinavia
INSERT INTO tax_rates (country, inclusive, percentage)
VALUES
('SE', TRUE, 2500), -- Sweden
('NO', TRUE, 2500), -- Norway
('DK', TRUE, 2500), -- Denmark
('IS', TRUE, 2400), -- Iceland
('FI', TRUE, 2400); -- Finland

-- Western Europe
INSERT INTO tax_rates (country, inclusive, percentage)
VALUES
('GB', TRUE, 2000),  -- United Kingdom
('CH', TRUE, 770),   -- Switzerland
('LI', TRUE, 770),   -- Liechtenstein
('DE', TRUE, 1900),  -- Germany
('FR', TRUE, 2000),  -- France
('IT', TRUE, 2200),  -- Italy
('ES', TRUE, 2100),  -- Spain
('NL', TRUE, 2100),  -- Netherlands
('BE', TRUE, 2100),  -- Belgium
('AT', TRUE, 2000),  -- Austria
('LU', TRUE, 1600),  -- Luxembourg
('IE', TRUE, 2300),  -- Ireland
('PT', TRUE, 2300),  -- Portugal

-- Central Europe
('PL', TRUE, 2300),  -- Poland
('CZ', TRUE, 2100),  -- Czech Republic
('SK', TRUE, 2000),  -- Slovakia
('HU', TRUE, 2700),  -- Hungary (highest in EU)
('SI', TRUE, 2200),  -- Slovenia
('HR', TRUE, 2500),  -- Croatia

-- Eastern Europe & Balkans
('RO', TRUE, 1900),  -- Romania
('BG', TRUE, 2000),  -- Bulgaria
('EE', TRUE, 2000),  -- Estonia
('LV', TRUE, 2100),  -- Latvia
('LT', TRUE, 2100),  -- Lithuania

-- Southeast & neighboring countries
('RS', TRUE, 2000),  -- Serbia
('BA', TRUE, 1700),  -- Bosnia & Herzegovina
('MK', TRUE, 1800),  -- North Macedonia
('MD', TRUE, 2000),  -- Moldova
('UA', TRUE, 2000),  -- Ukraine

-- Asia Pacific
('JP', TRUE, 1000),  -- Japan
('SG', TRUE, 900),   -- Singapore
('KR', TRUE, 1000),  -- South Korea
('IN', TRUE, 1800),  -- India: ~18% GST average
('CN', TRUE, 1300),  -- China

-- Middle East
('AE', TRUE, 500),   -- UAE
('SA', TRUE, 1500),  -- Saudi Arabia
('IL', TRUE, 1700),  -- Israel

-- Africa
('ZA', TRUE, 1500),  -- South Africa
('NG', TRUE, 750),   -- Nigeria

-- Latin America
('BR', TRUE, 1700),  -- Brazil
('MX', TRUE, 1600),  -- Mexico
('AR', TRUE, 2100),  -- Argentina
('CL', TRUE, 1900),  -- Chile

-- Oceania
('AU', TRUE, 1000),  -- Australia
('NZ', TRUE, 1500);  -- New Zealand
