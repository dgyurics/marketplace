-- Drop the existing foreign key constraint
ALTER TABLE addresses 
DROP CONSTRAINT addresses_user_id_fkey;

-- Add the new foreign key constraint with CASCADE
ALTER TABLE addresses 
ADD CONSTRAINT addresses_user_id_fkey 
FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;

-- Drop the existing foreign key constraints
ALTER TABLE orders 
DROP CONSTRAINT orders_user_id_fkey;

ALTER TABLE orders 
DROP CONSTRAINT orders_address_id_fkey;

-- Add the new foreign key constraints with CASCADE
ALTER TABLE orders 
ADD CONSTRAINT orders_user_id_fkey 
FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;

ALTER TABLE orders 
ADD CONSTRAINT orders_address_id_fkey 
FOREIGN KEY (address_id) REFERENCES addresses (id) ON DELETE CASCADE;

-- Drop the existing foreign key constraints
ALTER TABLE offers 
DROP CONSTRAINT offers_user_id_fkey;

-- Add the new foreign key constraints with CASCADE
ALTER TABLE offers 
ADD CONSTRAINT offers_user_id_fkey 
FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;