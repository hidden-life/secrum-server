ALTER TABLE key_bundles
    ADD COLUMN IF NOT EXISTS signed_prekey_sig TEXT;