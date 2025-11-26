CREATE TABLE sync_events (
     id         BIGSERIAL PRIMARY KEY,
     user_id    UUID        NOT NULL,
     type       TEXT        NOT NULL,
     payload    JSONB       NOT NULL,
     created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sync_events_user_id_id
    ON sync_events (user_id, id);