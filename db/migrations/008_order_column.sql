--payment_method is NULL when order_status is 'draft'
ALTER TABLE orders ALTER COLUMN payment_method DROP DEFAULT;
ALTER TABLE orders ALTER COLUMN payment_method DROP NOT NULL;

--updating order_status enum/column to support new value 'draft'
DROP INDEX IF EXISTS idx_orders_user_id_unique;
ALTER TABLE orders ALTER COLUMN status DROP DEFAULT;
ALTER TABLE orders ALTER COLUMN status TYPE TEXT;

DROP TYPE order_status_enum;
CREATE TYPE order_status_enum AS ENUM (
    'draft',
    'pending',
    'paid',
    'shipped',
    'delivered',
    'canceled',
    'refunded'
);

ALTER TABLE orders ALTER COLUMN status TYPE order_status_enum USING status::order_status_enum;
ALTER TABLE orders ALTER COLUMN status SET DEFAULT 'draft';
ALTER TABLE orders ALTER COLUMN status SET NOT NULL;

CREATE UNIQUE INDEX idx_orders_user_id_unique
ON orders (user_id)
WHERE status = 'draft' OR (status = 'pending' AND payment_method = 'stripe');