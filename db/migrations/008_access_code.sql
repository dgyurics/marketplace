CREATE TABLE access_codes (
  user_id BIGINT NOT NULL,
  code CHAR(6) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  UNIQUE (user_id),
  UNIQUE (code)
);

CREATE INDEX idx_access_codes_code ON access_codes (code);
