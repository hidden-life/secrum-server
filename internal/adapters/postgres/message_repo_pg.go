package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/message"
	"github.com/hidden-life/secrum-server/internal/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageRepositoryPG struct {
	pool *pgxpool.Pool
}

func NewMessageRepository(pool *pgxpool.Pool) ports.MessageRepository {
	return &MessageRepositoryPG{pool: pool}
}

func (m *MessageRepositoryPG) Save(ctx context.Context, msg *message.Message) error {
	const q = `INSERT INTO messages (id, sender_user_id, sender_device_id, recipient_user_id, recipient_device_id, ciphertext, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := m.pool.Exec(ctx, q,
		msg.ID,
		msg.SenderUserID,
		msg.SenderDeviceID,
		msg.RecipientUserID,
		msg.RecipientDeviceID,
		msg.CipherText,
		msg.CreatedAt,
	)

	return err
}

func (m *MessageRepositoryPG) GetPendingByRecipientDevice(ctx context.Context, deviceID uuid.UUID, limit int) ([]*message.Message, error) {
	const q = `
SELECT * FROM messages
WHERE recipient_device_id = $1 AND delivered_at IS NULL
ORDER BY created_at ASC
LIMIT $2`

	rows, err := m.pool.Query(ctx, q, deviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var output []*message.Message
	for rows.Next() {
		var m message.Message
		err := rows.Scan(&m.ID, &m.SenderUserID, &m.SenderDeviceID, &m.RecipientUserID, &m.RecipientDeviceID, &m.CipherText, &m.CreatedAt, &m.DeliveredAt, &m.ReadAt)
		if err != nil {
			return nil, err
		}

		output = append(output, &m)
	}

	return output, rows.Err()
}

func (m *MessageRepositoryPG) MarkDelivered(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	const q = `UPDATE messages SET delivered_at = $2 WHERE id = ANY($1) AND delivered_at IS NULL`
	now := time.Now().UTC()
	_, err := m.pool.Exec(ctx, q, ids, now)

	return err
}
