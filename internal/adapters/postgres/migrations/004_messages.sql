CREATE TABLE IF NOT EXISTS messages (
                                        id UUID PRIMARY KEY,
                                        sender_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sender_device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    recipient_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipient_device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    ciphertext TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    delivered_at TIMESTAMPTZ,
    read_at TIMESTAMPTZ
    );

CREATE INDEX IF NOT EXISTS idx_messages_recipient_device_created
    ON messages (recipient_device_id, created_at);

CREATE INDEX IF NOT EXISTS idx_messages_sender_device_created
    ON messages (sender_device_id, created_at);