ALTER TABLE orders
    ALTER COLUMN status TYPE TEXT;

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