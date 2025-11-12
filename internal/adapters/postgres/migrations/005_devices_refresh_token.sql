UPDATE devices SET name = COALESCE(NULLIF(name, ''), 'unknown');
UPDATE devices SET platform = COALESCE(NULLIF(platform, ''), 'unknown');

ALTER TABLE devices
    ADD COLUMN IF NOT EXISTS refresh_token_hash TEXT,
    ADD COLUMN IF NOT EXISTS refresh_expires_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_devices_refresh_expires_at
    ON devices (refresh_expires_at);