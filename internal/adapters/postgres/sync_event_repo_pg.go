package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SyncEventRepositoryPG struct {
	pool *pgxpool.Pool
}

func NewSyncEventRepository(pool *pgxpool.Pool) *SyncEventRepositoryPG {
	return &SyncEventRepositoryPG{pool: pool}
}

func (s *SyncEventRepositoryPG) Append(ctx context.Context, userID uuid.UUID, eventType string, payload any) (int64, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	const q = `INSERT INTO sync_events (user_id, type, payload) VALUES ($1, $2, $3) RETURNING id`

	var id int64
	if err := s.pool.QueryRow(ctx, q, userID, eventType, data).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (s *SyncEventRepositoryPG) ListSince(ctx context.Context, userID uuid.UUID, after int64, limit int) ([]ports.SyncEvent, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	const q = `SELECT id, user_id, type, payload, created_at FROM sync_events WHERE user_id = $1 AND id > $2 ORDER BY id LIMIT $3`

	rows, err := s.pool.Query(ctx, q, userID, after, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ports.SyncEvent
	for rows.Next() {
		var e ports.SyncEvent
		if err := rows.Scan(&e.ID, &e.UserID, &e.Type, &e.Payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, e)
	}

	return res, rows.Err()
}

func (s *SyncEventRepositoryPG) GetLastID(ctx context.Context, userID uuid.UUID) (int64, error) {
	const q = `SELECT COALESCE(MAX(id), 0) FROM sync_events WHERE user_id = $1`
	var id int64
	if err := s.pool.QueryRow(ctx, q, userID).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}
