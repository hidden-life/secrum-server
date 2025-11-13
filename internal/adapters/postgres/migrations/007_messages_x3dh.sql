ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS x3dh_otpk_id UUID NULL;

CREATE INDEX IF NOT EXISTS idx_messages_x3dh_otpk_id
    ON messages (x3dh_otpk_id);