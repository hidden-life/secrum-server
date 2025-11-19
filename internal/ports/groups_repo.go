package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/group"
)

type GroupRepository interface {
	Create(context.Context, *group.Group) error
	GetByID(context.Context, uuid.UUID) (*group.Group, error)
	ListByUser(context.Context, uuid.UUID) ([]*group.Group, error)
	Update(context.Context, *group.Group) error
}

type GroupMemberRepository interface {
	AddMember(context.Context, *group.Member) error
	RemoveMember(context.Context, uuid.UUID, uuid.UUID) error
	List(context.Context, uuid.UUID) ([]*group.Member, error)
	IsMember(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	GetRole(context.Context, uuid.UUID, uuid.UUID) (group.MemberRole, error)
}
