ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS ephemeral_pub_key TEXT NULL;