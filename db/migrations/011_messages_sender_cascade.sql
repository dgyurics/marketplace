ALTER TABLE messages
DROP CONSTRAINT messages_sender_id_fkey,
ADD CONSTRAINT messages_sender_id_fkey
FOREIGN KEY (sender_id) REFERENCES users (id) ON DELETE CASCADE;
