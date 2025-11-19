package group

import (
	"time"

	"github.com/google/uuid"
)

type MemberRole string

const (
	RoleOwner  MemberRole = "owner"
	RoleAdmin  MemberRole = "admin"
	RoleMember MemberRole = "member"
)

type Group struct {
	ID        uuid.UUID
	Name      string
	AvatarURL *string

	CreatedBy uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	IsActive  bool
}

type Member struct {
	GroupID  uuid.UUID
	UserID   uuid.UUID
	Role     MemberRole
	JoinedAt time.Time
	IsActive bool
}

func NewGroup(name string, createdBy uuid.UUID, avatar *string) *Group {
	now := time.Now().UTC()
	return &Group{
		ID:        uuid.New(),
		Name:      name,
		CreatedBy: createdBy,
		AvatarURL: avatar,
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
	}
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
