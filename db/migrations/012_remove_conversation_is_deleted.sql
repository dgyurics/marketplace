DROP INDEX idx_conversations_updated_at_is_deleted;

ALTER TABLE conversations DROP COLUMN is_deleted;
