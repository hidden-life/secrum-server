CREATE TABLE IF NOT EXISTS one_time_prekeys (
    id              UUID PRIMARY KEY,
    device_id       UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    public_key      TEXT NOT NULL,         -- X25519 pub
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at         TIMESTAMPTZ NULL       -- mark on use
    );

CREATE INDEX IF NOT EXISTS idx_otpk_device_id ON one_time_prekeys (device_id);
CREATE INDEX IF NOT EXISTS idx_otpk_unused ON one_time_prekeys (device_id, used_at);