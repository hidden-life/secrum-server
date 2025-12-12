ALTER TABLE devices
    ADD COLUMN identity_key TEXT,
    ADD COLUMN signed_prekey TEXT,
    ADD COLUMN signed_prekey_signature TEXT;