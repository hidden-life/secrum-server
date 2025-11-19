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

	CreatedBy uuid.UUID // it's owner!
	CreatedAt time.Time
	UpdatedAt time.Time
	IsActive  bool
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

func (d *Group) RenameGroup(name string) {
	d.Name = name
	d.UpdatedAt = time.Now().UTC()
}

func (d *Group) DeactivateGroup() {
	d.IsActive = false
	d.UpdatedAt = time.Now().UTC()
}
