ALTER TABLE key_bundles
    ADD CONSTRAINT key_bundles_device_id_key UNIQUE (device_id);