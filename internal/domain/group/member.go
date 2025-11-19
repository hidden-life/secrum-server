package group

import (
	"time"

	"github.com/google/uuid"
)

type Member struct {
	GroupID   uuid.UUID
	UserID    uuid.UUID
	Role      MemberRole
	JoinedAt  time.Time
	RemovedAt *time.Time
	IsActive  bool
}

func NewMember(groupID, userID uuid.UUID, role MemberRole) *Member {
	return &Member{
		GroupID:  groupID,
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now().UTC(),
		IsActive: true,
	}
}

func (m *Member) DeactivateMember() {
	now := time.Now().UTC()
	m.IsActive = false
	m.JoinedAt = now
}
