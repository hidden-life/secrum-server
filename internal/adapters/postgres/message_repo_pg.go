package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/message"
	"github.com/hidden-life/secrum-server/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageRepositoryPG struct {
	pool *pgxpool.Pool
}

func NewMessageRepository(pool *pgxpool.Pool) ports.MessageRepository {
	return &MessageRepositoryPG{pool: pool}
}

func (m *MessageRepositoryPG) Save(ctx context.Context, msg *message.Message) error {
	const q = `INSERT INTO messages (
                      id, 
                      sender_user_id, 
                      sender_device_id, 
                      recipient_user_id,
                      recipient_device_id, 
                      ciphertext, 
                      x3dh_otpk_id, 
                      ephemeral_pub_key,
                      created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := m.pool.Exec(ctx, q,
		msg.ID,
		msg.SenderUserID,
		msg.SenderDeviceID,
		msg.RecipientUserID,
		msg.RecipientDeviceID,
		msg.CipherText,
		msg.X3DHOTPKID,
		msg.PubKey,
		msg.CreatedAt,
	)

	return err
}

func (m *MessageRepositoryPG) GetPendingByRecipientDevice(ctx context.Context, deviceID uuid.UUID, limit int) ([]*message.Message, error) {
	const q = `
SELECT id, 
       sender_user_id, 
       sender_device_id, 
       recipient_user_id, 
       recipient_device_id, 
       ciphertext,
       created_at,
       delivered_at,
       read_at,
       x3dh_otpk_id,
       ephemeral_pub_key
FROM messages
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
		err := rows.Scan(
			&m.ID,
			&m.SenderUserID,
			&m.SenderDeviceID,
			&m.RecipientUserID,
			&m.RecipientDeviceID,
			&m.CipherText,
			&m.CreatedAt,
			&m.DeliveredAt,
			&m.ReadAt,
			&m.X3DHOTPKID,
			&m.PubKey,
		)
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

func (m *MessageRepositoryPG) MarkRead(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	const q = `UPDATE messages SET read_at = $2 WHERE id = ANY($1) AND read_at IS NULL`
	now := time.Now().UTC()
	_, err := m.pool.Exec(ctx, q, ids, now)

	return err
}

func (m *MessageRepositoryPG) UserChatsList(ctx context.Context, userID uuid.UUID) ([]ports.ChatSummary, error) {
	const q = `WITH user_messages AS (
		SELECT 
			CASE 
				WHEN sender_user_id = $1 THEN recipient_user_id
				ELSE sender_user_id
			END AS peer_user_id,
			id,
			ciphertext,
			created_at,
			recipient_user_id,
			read_at
		FROM messages
		WHERE sender_user_id = $1 OR recipient_user_id = $1
	), agg AS (
		SELECT 
			peer_user_id, 
			MAX(created_at) AS last_message_at,
			(ARRAY_AGG(ciphertext ORDER BY created_at DESC))[1] AS last_cipher_text,
			COUNT(*) FILTER (WHERE recipient_user_id = $1 AND read_at IS NULL) AS unread_count
		FROM user_messages
		GROUP BY peer_user_id
	)
	SELECT peer_user_id, last_message_at, last_cipher_text, unread_count FROM agg ORDER BY last_message_at DESC;
`

	rows, err := m.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ports.ChatSummary
	for rows.Next() {
		var cs ports.ChatSummary
		if err := rows.Scan(&cs.PeerUserID, &cs.LastMessageAt, &cs.LastCipherText, &cs.UnreadCount); err != nil {
			return nil, err
		}

		res = append(res, cs)
	}

	return res, rows.Err()
}

func (m *MessageRepositoryPG) SaveMany(ctx context.Context, messages []*message.Message) error {
	if len(messages) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	const q = `INSERT INTO messages (
					  id, 
					  sender_user_id, 
					  sender_device_id,
                      recipient_user_id,
                      recipient_device_id,
                      group_id,
                      ciphertext,
                      created_at,
                      x3dh_otpk_id,
                      ephemeral_pub_key
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	for _, msg := range messages {
		batch.Queue(q,
			msg.ID,
			msg.SenderUserID,
			msg.SenderDeviceID,
			msg.RecipientUserID,
			msg.RecipientDeviceID,
			msg.GroupID,
			msg.CipherText,
			msg.CreatedAt,
			msg.X3DHOTPKID,
			msg.PubKey)
	}

	res := m.pool.SendBatch(ctx, batch)
	defer res.Close()

	for i := 0; i < len(messages); i++ {
		if _, err := res.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (m *MessageRepositoryPG) GetGroupMessages(ctx context.Context, gid uuid.UUID, limit int, before *time.Time) ([]*message.Message, error) {
	q := `SELECT 
			id, 
			sender_user_id,
			sender_device_id,
			recipient_user_id,
			recipient_device_id,
			group_id,
			ciphertext,
			created_at,
			delivered_at,
			read_at,
			x3dh_otpk_id,
			ephemeral_pub_key
		FROM messages
		WHERE group_id = $1`

	args := []interface{}{gid}
	argPos := 2
	if before != nil {
		q += fmt.Sprintf(" AND created_at < $%d", argPos)
		args = append(args, *before)
		argPos++
	}

	q += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argPos)
	args = append(args, limit)

	rows, err := m.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*message.Message
	for rows.Next() {
		var msg message.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.SenderUserID,
			&msg.SenderDeviceID,
			&msg.RecipientUserID,
			&msg.RecipientDeviceID,
			&msg.GroupID,
			&msg.CipherText,
			&msg.CreatedAt,
			&msg.DeliveredAt,
			&msg.ReadAt,
			&msg.X3DHOTPKID,
			&msg.PubKey); err != nil {
			return nil, err
		}
		result = append(result, &msg)
	}

	return result, nil
}

func (m *MessageRepositoryPG) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*message.Message, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	const q = `SELECT
		id,
		sender_user_id,
		sender_device_id,
		recipient_user_id,
		recipient_device_id,
		ciphertext,
		created_at,
		delivered_at,
		read_at,
		x3dh_otpk_id,
		ephemeral_pub_key,
		group_id
		FROM messages
		WHERE id = ANY($1)`

	rows, err := m.pool.Query(ctx, q, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*message.Message
	for rows.Next() {
		var msg message.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.SenderUserID,
			&msg.SenderDeviceID,
			&msg.RecipientUserID,
			&msg.RecipientDeviceID,
			&msg.CipherText,
			&msg.CreatedAt,
			&msg.DeliveredAt,
			&msg.ReadAt,
			&msg.X3DHOTPKID,
			&msg.PubKey,
			&msg.GroupID,
		); err != nil {
			return nil, err
		}

		res = append(res, &msg)
	}

	return res, rows.Err()
}

func (m *MessageRepositoryPG) GetChatHistory(ctx context.Context, userID, peerID uuid.UUID, limit int, before *time.Time) ([]*message.Message, error) {
	q := `SELECT
		id,
		sender_user_id,
		sender_device_id,
		recipient_user_id,
		recipient_device_id,
		ciphertext,
		created_at,
		delivered_at,
		read_at,
		x3dh_otpk_id,
		ephemeral_pub_key,
		group_id
		FROM messages
		WHERE (
		    sender_user_id = $1 AND recipient_user_id = $2
		) OR (
		    sender_user_id = $2 AND recipient_user_id = $1
		)`

	args := []any{userID, peerID}
	idx := 3

	if before != nil {
		q += fmt.Sprintf(" AND created_at < $%d", idx)
		args = append(args, *before)
		idx++
	}

	q += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", idx)
	args = append(args, limit)

	rows, err := m.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*message.Message
	for rows.Next() {
		var msg message.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.SenderUserID,
			&msg.SenderDeviceID,
			&msg.RecipientUserID,
			&msg.RecipientDeviceID,
			&msg.CipherText,
			&msg.CreatedAt,
			&msg.DeliveredAt,
			&msg.ReadAt,
			&msg.X3DHOTPKID,
			&msg.PubKey,
			&msg.GroupID,
		); err != nil {
			return nil, err
		}

		result = append(result, &msg)
	}

	return result, rows.Err()
}
