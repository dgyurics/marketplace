DROP TABLE CLAIMS;

CREATE TYPE purchase_intent_status_enum AS ENUM (
    'pending',
    'accepted',
    'rejected',
    'canceled',
    'completed'
);
CREATE TABLE purchase_intents (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  product_id BIGINT NOT NULL,
  offer_price BIGINT NOT NULL,
  status purchase_intent_status_enum NOT NULL,
  pickup_notes TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
  FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT
);
