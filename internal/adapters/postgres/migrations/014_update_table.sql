CREATE TABLE groups (
                        id             uuid                     NOT NULL PRIMARY KEY,
                        owner_user_id  uuid                     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                        title          text                     NOT NULL,
                        created_at     timestamp with time zone NOT NULL,
                        updated_at     timestamp with time zone NOT NULL,
                        is_active      boolean                  NOT NULL DEFAULT true
);

ALTER TABLE groups OWNER TO secrum_user;

CREATE INDEX idx_groups_owner ON groups (owner_user_id);
CREATE INDEX idx_groups_active ON groups (is_active);

-- Много-ко-многим: группы ↔ пользователи
CREATE TABLE group_members (
                               id         uuid                     NOT NULL PRIMARY KEY,
                               group_id   uuid                     NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
                               user_id    uuid                     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                               role       varchar(16)              NOT NULL, -- 'admin' | 'member'
                               added_at   timestamp with time zone NOT NULL,
                               removed_at timestamp with time zone,
                               is_active  boolean                  NOT NULL DEFAULT true
);

ALTER TABLE group_members OWNER TO secrum_user;

-- один активный membership на пользователя в группе
CREATE UNIQUE INDEX uq_group_members_active
    ON group_members (group_id, user_id)
    WHERE is_active = true;

CREATE INDEX idx_group_members_group ON group_members (group_id);
CREATE INDEX idx_group_members_user ON group_members (user_id);