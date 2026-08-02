ALTER TABLE orders ADD COLUMN idempotency_key VARCHAR(64);

CREATE UNIQUE INDEX idx_orders_idempotency_key
ON orders (idempotency_key)
WHERE idempotency_key IS NOT NULL;
