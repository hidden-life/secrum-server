ALTER TABLE users
    ADD COLUMN allowed_mime_types TEXT[] DEFAULT NULL;