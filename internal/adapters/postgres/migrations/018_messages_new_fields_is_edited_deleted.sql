ALTER TABLE messages
    ADD COLUMN is_edited boolean DEFAULT false,
    ADD COLUMN is_deleted boolean DEFAULT false;