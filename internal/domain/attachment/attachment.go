package attachment

import (
	"time"

	"github.com/google/uuid"
)

type Attachment struct {
	ID         uuid.UUID
	UploadedBy uuid.UUID
	BlobPath   string
	CreatedAt  time.Time
	FileSize   *int64
	MimeType   *string
	Sha256Hex  *string
	IsDeleted  bool
}

func New(uploaderID uuid.UUID, blobPath string, fileSize *int64, mimeType *string, sha *string) *Attachment {
	return &Attachment{
		ID:         uuid.New(),
		UploadedBy: uploaderID,
		BlobPath:   blobPath,
		CreatedAt:  time.Now().UTC(),
		FileSize:   fileSize,
		MimeType:   mimeType,
		Sha256Hex:  sha,
		IsDeleted:  false,
	}
}
