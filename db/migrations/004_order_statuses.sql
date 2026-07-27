ALTER TABLE orders
    ALTER COLUMN status DROP DEFAULT;

ALTER TABLE orders
    ALTER COLUMN status TYPE TEXT;

DROP INDEX IF EXISTS idx_orders_user_id_unique;

DROP TYPE order_status_enum;

CREATE TYPE order_status_enum AS ENUM (
    'pending',
    'paid',
    'shipped',
    'delivered',
    'canceled',
    'refunded'
);

ALTER TABLE orders
    ALTER COLUMN status TYPE order_status_enum
    USING status::order_status_enum;

ALTER TABLE orders
    ALTER COLUMN status SET DEFAULT 'pending';

CREATE UNIQUE INDEX idx_orders_user_id_unique
ON orders (user_id)
WHERE status = 'pending';