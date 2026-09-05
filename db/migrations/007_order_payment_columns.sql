CREATE TYPE payment_method_enum AS ENUM ('stripe', 'on_delivery');

CREATE TYPE payment_status_enum AS ENUM ('unpaid', 'paid', 'refunded');

ALTER TABLE orders
ADD COLUMN IF NOT EXISTS payment_method payment_method_enum;

ALTER TABLE orders
ADD COLUMN IF NOT EXISTS payment_status payment_status_enum;

UPDATE orders
SET payment_method = 'stripe'
WHERE payment_method IS NULL;

UPDATE orders
SET payment_status = CASE
    WHEN status = 'paid' THEN 'paid'::payment_status_enum
    WHEN status = 'refunded' THEN 'refunded'::payment_status_enum
    ELSE 'unpaid'::payment_status_enum
END
WHERE payment_status IS NULL;

ALTER TABLE orders
ALTER COLUMN payment_method SET DEFAULT 'stripe',
ALTER COLUMN payment_method SET NOT NULL,
ALTER COLUMN payment_status SET DEFAULT 'unpaid',
ALTER COLUMN payment_status SET NOT NULL;
