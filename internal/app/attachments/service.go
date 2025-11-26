package attachments

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/attachment"
	"github.com/hidden-life/secrum-server/internal/ports"
	"go.uber.org/zap"
)

type Service struct {
	log                  *zap.Logger
	attachmentRepository ports.AttachmentRepository
	storage              ports.FileStorage
	basePath             string
	maxSize              int64
}

type UploadResultResponse struct {
	ID       string  `json:"id"`
	FileSize *int64  `json:"file_size,omitempty"`
	MimeType *string `json:"mime_type,omitempty"`
}

type DownloadInfo struct {
	ID       string
	MimeType *string
	Size     *int64
	Reader   io.ReadCloser
}

func NewService(log *zap.Logger, repo ports.AttachmentRepository, s ports.FileStorage, basePath string, maxSize int64) *Service {
	if maxSize <= 0 {
		maxSize = 1024 * 1024 * 50 // 50 MB as default size
	}

	return &Service{
		log:                  log,
		attachmentRepository: repo,
		storage:              s,
		basePath:             basePath,
		maxSize:              maxSize,
	}
}

func (s *Service) Upload(ctx context.Context, uploaderID string, r io.Reader, fileSize *int64, mimeType *string) (*UploadResultResponse, error) {
	uid, err := uuid.Parse(uploaderID)
	if err != nil {
		return nil, fmt.Errorf("invalid uploader user id: %s", uploaderID)
	}

	attID := uuid.New()
	blobPath := s.buildBlobPath(uid, attID)

	var limitReader io.Reader = r
	if s.maxSize > 0 {
		limitReader = io.LimitReader(r, s.maxSize+1)
	}

	h := sha256.New()
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		_, err := io.Copy(io.MultiWriter(pw, h), limitReader)
		if err != nil {
			s.log.Warn("failed to read upload stream", zap.Error(err))
		}
	}()

	if err := s.storage.Save(ctx, blobPath, pr); err != nil {
		return nil, fmt.Errorf("failed to save attachment blob: %w", err)
	}

	sum := h.Sum(nil)
	shaHex := hex.EncodeToString(sum)
	now := time.Now().UTC()
	a := &attachment.Attachment{
		ID:         attID,
		UploadedBy: uid,
		BlobPath:   blobPath,
		CreatedAt:  now,
		MimeType:   mimeType,
		FileSize:   fileSize,
		Sha256Hex:  &shaHex,
		IsDeleted:  false,
	}

	if err := s.attachmentRepository.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("failed to save attachment metadta: %w", err)
	}

	s.log.Info("successfully uploaded attachment",
		zap.String("attachment_id", attID.String()),
		zap.String("blob_path", blobPath),
		zap.String("uploader", uid.String()))

	return &UploadResultResponse{
		ID:       attID.String(),
		FileSize: fileSize,
		MimeType: mimeType,
	}, nil
}

func (s *Service) Download(ctx context.Context, userID, attachmentID string) (*DownloadInfo, error) {
	if userID == "" {
		return nil, fmt.Errorf("empty user id")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}

	id, err := uuid.Parse(attachmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid attachment id: %s", attachmentID)
	}

	a, err := s.attachmentRepository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load attachment: %w", err)
	}

	if a == nil || a.IsDeleted {
		return nil, fmt.Errorf("attachment does not exist")
	}

	if a.UploadedBy != uid {
		return nil, fmt.Errorf("access forbidden")
	}

	rc, err := s.storage.Open(ctx, a.BlobPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open attachment blob: %w", err)
	}

	return &DownloadInfo{
		ID:       a.ID.String(),
		MimeType: a.MimeType,
		Size:     a.FileSize,
		Reader:   rc,
	}, nil
}

func (s *Service) buildBlobPath(userID, attachmentID uuid.UUID) string {
	return filepath.Join("attachments", userID.String(), attachmentID.String()+".bin")
}
