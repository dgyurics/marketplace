CREATE TYPE conversation_type_enum AS ENUM ('support', 'notification');

CREATE TABLE conversations (
    id BIGINT PRIMARY KEY,
    type conversation_type_enum NOT NULL,
    subject TEXT NOT NULL,
    last_message_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    recipient_id BIGINT NOT NULL,
    recipient_last_read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    FOREIGN KEY (recipient_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_conversations_recipient_id_last_message_at_is_deleted ON conversations(recipient_id, last_message_at DESC)
WHERE is_deleted = false;

CREATE TABLE messages (
    id BIGINT PRIMARY KEY,
    sender_id BIGINT,
    conversation_id BIGINT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

CREATE INDEX idx_messages_conversation_id ON messages(conversation_id, created_at DESC);
