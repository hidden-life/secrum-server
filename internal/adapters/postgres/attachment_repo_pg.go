package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/attachment"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AttachmentRepositoryPG struct {
	pool *pgxpool.Pool
}

func NewAttachmentRepository(p *pgxpool.Pool) *AttachmentRepositoryPG {
	return &AttachmentRepositoryPG{pool: p}
}

func (r *AttachmentRepositoryPG) Create(ctx context.Context, attachment *attachment.Attachment) error {
	const q = `INSERT INTO attachments(
                        id, 
                        uploader_user_id, 
                        blob_path, 
                        created_at, 
                        file_size, 
                        mime_type, 
                        sha256_hex, 
                        is_deleted)
                        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, q,
		attachment.ID,
		attachment.UploadedBy,
		attachment.BlobPath,
		attachment.CreatedAt,
		attachment.FileSize,
		attachment.MimeType,
		attachment.Sha256Hex,
		attachment.IsDeleted)

	return err
}

func (r *AttachmentRepositoryPG) GetByID(ctx context.Context, id uuid.UUID) (*attachment.Attachment, error) {
	const q = `SELECT id, uploader_user_id, blob_path, created_at, file_size, mime_type, sha256_hex, is_deleted FROM attachments WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)
	var a attachment.Attachment
	if err := row.Scan(&a.ID, &a.UploadedBy, &a.BlobPath, &a.CreatedAt, &a.FileSize, &a.MimeType, &a.Sha256Hex, &a.IsDeleted); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &a, nil
}

func (r *AttachmentRepositoryPG) MarDeleted(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE attachments SET is_deleted = TRUE WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id)
	return err
}
