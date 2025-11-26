CREATE TABLE chat_user_state (
     user_id uuid NOT NULL,
     peer_user_id uuid NOT NULL,

     pinned boolean DEFAULT false,
     archived boolean DEFAULT false,
     muted boolean DEFAULT false,

     updated_at timestamptz DEFAULT now(),

     PRIMARY KEY (user_id, peer_user_id),
     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
     FOREIGN KEY (peer_user_id) REFERENCES users(id) ON DELETE CASCADE
);