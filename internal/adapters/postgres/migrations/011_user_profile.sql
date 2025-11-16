ALTER TABLE users
    ADD COLUMN IF NOT EXISTS display_name      VARCHAR(128)          NULL,
    ADD COLUMN IF NOT EXISTS avatar_url        TEXT                  NULL,
    ADD COLUMN IF NOT EXISTS status_message    TEXT                  NULL,
    ADD COLUMN IF NOT EXISTS safety_fingerprint VARCHAR(255)         NULL;