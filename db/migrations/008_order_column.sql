--payment_method is NULL when order_status is 'created'
ALTER TABLE orders ALTER COLUMN payment_method DROP DEFAULT;
ALTER TABLE orders ALTER COLUMN payment_method DROP NOT NULL;

--updating order_status enum/column to support new value 'created'
ALTER TABLE orders ALTER COLUMN status DROP DEFAULT;
ALTER TABLE orders ALTER COLUMN status TYPE TEXT;

DROP INDEX IF EXISTS idx_orders_user_id_unique;
DROP TYPE order_status_enum;
CREATE TYPE order_status_enum AS ENUM (
    'created',
    'pending',
    'paid',
    'shipped',
    'delivered',
    'canceled',
    'refunded'
);

ALTER TABLE orders ALTER COLUMN status TYPE order_status_enum USING status::order_status_enum;
ALTER TABLE orders ALTER COLUMN status SET DEFAULT 'created';
ALTER TABLE orders ALTER COLUMN status SET NOT NULL;

CREATE UNIQUE INDEX idx_orders_user_id_unique
ON orders (user_id)
WHERE status = 'created' OR (status = 'pending' AND payment_method = 'stripe');