package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/key"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OTPKRepository struct {
	pool *pgxpool.Pool
}

func NewOTPKRepository(pool *pgxpool.Pool) *OTPKRepository {
	return &OTPKRepository{pool: pool}
}

func (r *OTPKRepository) BulkInsert(ctx context.Context, keys []*key.OneTimePreKey) error {
	batch := &pgx.Batch{}
	const q = `INSERT INTO one_time_prekeys (id, device_id, public_key, created_at, used_at) VALUES ($1, $2, $3, $4, $5)`
	for _, k := range keys {
		batch.Queue(q, k.ID, k.DeviceID, k.PublicKey, k.CreatedAt, k.UsedAt)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	_, err := br.Exec()
	return err
}

func (r *OTPKRepository) GetOneUnusedAndMarkUsed(ctx context.Context, id uuid.UUID) (*key.OneTimePreKey, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// find any unused key
	const q = `SELECT id, public_key, created_at, used_at 
		FROM one_time_prekeys 
		WHERE device_id = $1 AND used_at IS NULL
		ORDER BY created_at ASC
		FOR UPDATE SKIP LOCKED LIMIT 1`

	row := tx.QueryRow(ctx, q, id)
	var k key.OneTimePreKey
	k.DeviceID = id
	if err := row.Scan(&k.ID, &k.PublicKey, &k.CreatedAt, &k.UsedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	// mark as used
	const updQ = `UPDATE one_time_prekeys SET used_at = NOW() WHERE id = $1`
	if _, err := tx.Exec(ctx, updQ, k.ID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &k, nil
}
