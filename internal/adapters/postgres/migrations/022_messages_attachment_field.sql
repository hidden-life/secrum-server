ALTER TABLE messages
    ADD COLUMN attachment_id UUID REFERENCES attachments(id);