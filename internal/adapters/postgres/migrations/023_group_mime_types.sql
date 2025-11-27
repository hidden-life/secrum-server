ALTER TABLE groups
    ADD COLUMN allowed_mime_types TEXT[] DEFAULT NULL;