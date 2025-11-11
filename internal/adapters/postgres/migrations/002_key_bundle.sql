CREATE TABLE IF NOT EXISTS key_bundles (
    id UUID PRIMARY KEY,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    identity_key TEXT NOT NULL,
    signed_prekey TEXT NOT NULL,
    one_time_prekeys TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_keybundle_device_id ON key_bundles(device_id);