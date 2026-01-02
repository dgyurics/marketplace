ALTER TABLE users DROP COLUMN requires_setup;

CREATE TABLE IF NOT EXISTS registration_codes (
  code CHAR(6) PRIMARY KEY,
  user_id BIGINT NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- For cleanup queries
CREATE INDEX idx_registration_codes_expires ON registration_codes(expires_at);

ALTER TABLE users ADD COLUMN verified BOOLEAN NOT NULL DEFAULT false;

DROP TABLE pending_users;