ALTER TABLE messages
    -- Pin
    ADD COLUMN pinned_by UUID[] NOT NULL DEFAULT '{}',
    ADD COLUMN pinned_at TIMESTAMPTZ NULL,

    -- Forwarding / quoting
    ADD COLUMN forwarded_from_msg_id UUID NULL,
    ADD COLUMN forwarded_from_user_id UUID NULL,
    ADD COLUMN quoted_message_id UUID NULL,

    -- Media metadata (один медиа-объект на сообщение)
    ADD COLUMN has_media BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN media_mime_type TEXT NULL,
    ADD COLUMN media_size_bytes BIGINT NULL,
    ADD COLUMN media_duration_ms INT NULL,
    ADD COLUMN media_width INT NULL,
    ADD COLUMN media_height INT NULL,
    ADD COLUMN media_blurhash TEXT NULL;