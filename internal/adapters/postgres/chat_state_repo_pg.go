package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatStateRepositoryPG struct {
	pool *pgxpool.Pool
}

func NewChatStateRepository(pool *pgxpool.Pool) *ChatStateRepositoryPG {
	return &ChatStateRepositoryPG{pool: pool}
}

func (c *ChatStateRepositoryPG) SetPinned(ctx context.Context, userID, peerID uuid.UUID, isPinned bool) error {
	const q = `INSERT INTO chat_user_state (user_id, peer_user_id, pinned)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, peer_user_id) DO UPDATE
		SET pinned = EXCLUDED.pinned, updated_at = NOW()`

	_, err := c.pool.Exec(ctx, q, userID, peerID, isPinned)

	return err
}

func (c *ChatStateRepositoryPG) SetArchived(ctx context.Context, userID, peerID uuid.UUID, isArchived bool) error {
	const q = `INSERT INTO chat_user_state (user_id, peer_user_id, archived)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, peer_user_id) DO UPDATE
		SET archived = EXCLUDED.archived, updated_at = NOW()`

	_, err := c.pool.Exec(ctx, q, userID, peerID, isArchived)

	return err
}

func (c *ChatStateRepositoryPG) SetMuted(ctx context.Context, userID, peerID uuid.UUID, isMuted bool) error {
	const q = `INSERT INTO chat_user_state (user_id, peer_user_id, muted)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, peer_user_id) DO UPDATE
		SET muted = EXCLUDED.muted, updated_at = NOW()`

	_, err := c.pool.Exec(ctx, q, userID, peerID, isMuted)

	return err
}

func (c *ChatStateRepositoryPG) GetState(ctx context.Context, userID, peerID uuid.UUID) (*ports.ChatState, error) {
	const q = `SELECT pinned, archived, muted FROM chat_user_state WHERE user_id = $1 AND peer_user_id = $2`

	var state ports.ChatState
	err := c.pool.QueryRow(ctx, q, userID, peerID).Scan(&state.Pinned, &state.Archived, &state.Muted)
	if err != nil {
		return &ports.ChatState{}, err
	}

	return &state, nil
}
