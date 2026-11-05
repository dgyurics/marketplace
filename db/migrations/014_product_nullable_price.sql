-- Draft products may not have a price yet.
ALTER TABLE products ALTER COLUMN price DROP NOT NULL;

-- Published and archived products must always have a price.
ALTER TABLE products
ADD CONSTRAINT products_price_required
CHECK (status = 'draft' OR price IS NOT NULL);
