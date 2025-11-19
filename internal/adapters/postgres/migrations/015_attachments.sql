CREATE TABLE attachments (
     id               uuid PRIMARY KEY,
     uploader_user_id uuid                     NOT NULL
         REFERENCES users(id) ON DELETE CASCADE,
     blob_path        text                     NOT NULL,
     created_at       timestamptz              NOT NULL DEFAULT now(),
     file_size        bigint,
     mime_type        text,
     sha256_hex       char(64),
     is_deleted       boolean                  NOT NULL DEFAULT false
);

CREATE INDEX idx_attachments_uploader ON attachments (uploader_user_id);
CREATE INDEX idx_attachments_created  ON attachments (created_at);