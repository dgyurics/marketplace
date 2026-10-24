CREATE INDEX idx_users_guest_created_at
ON users (created_at)
WHERE role = 'guest';

CREATE INDEX idx_orders_user_id
ON orders (user_id);

CREATE INDEX idx_offers_user_id
ON offers (user_id);

CREATE INDEX idx_refresh_tokens_user_id_last_used
ON refresh_tokens (user_id, last_used);

CREATE INDEX idx_registration_codes_user_id
ON registration_codes (user_id);

CREATE INDEX idx_password_reset_codes_user_id
ON password_reset_codes (user_id);

CREATE INDEX idx_conversations_recipient_id
ON conversations (recipient_id);

CREATE INDEX idx_messages_sender_id
ON messages (sender_id);

CREATE INDEX idx_orders_address_id
ON orders (address_id);
