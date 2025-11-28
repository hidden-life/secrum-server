package groups

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/group"
	"github.com/hidden-life/secrum-server/internal/ports"
	"github.com/hidden-life/secrum-server/internal/real_time"
	"go.uber.org/zap"
)

type Service struct {
	log                   *zap.Logger
	groupRepository       ports.GroupRepository
	groupMemberRepository ports.GroupMemberRepository
	userRepository        ports.UserRepository
	deviceRepository      ports.DeviceRepository
	messageRepository     ports.MessageRepository
	hub                   *real_time.DeliveryHub
	syncRepository        ports.SyncEventRepository
}

type GroupResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Role      string  `json:"role"`
}

type CreateGroupRequest struct {
	Name      string   `json:"name"`
	AvatarURL *string  `json:"avatar_url,omitempty"`
	Members   []string `json:"members,omitempty"` // user id
}

type AddMemberRequest struct {
	UserID string `json:"user_id"`
}

type GroupMemberResponse struct {
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

func NewService(
	log *zap.Logger,
	groupRepo ports.GroupRepository,
	groupMemberRepo ports.GroupMemberRepository,
	userRepo ports.UserRepository,
	deviceRepo ports.DeviceRepository,
	messageRepo ports.MessageRepository,
	hub *real_time.DeliveryHub,
	syncRepository ports.SyncEventRepository,
) *Service {
	return &Service{
		log:                   log,
		groupRepository:       groupRepo,
		groupMemberRepository: groupMemberRepo,
		userRepository:        userRepo,
		deviceRepository:      deviceRepo,
		messageRepository:     messageRepo,
		hub:                   hub,
		syncRepository:        syncRepository,
	}
}

// CreateGroup creates a group where creator is owner of group
func (s *Service) CreateGroup(ctx context.Context, creator, name string, avatarURL *string, members []string) (*GroupResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("group name is required")
	}

	creatorUUID, err := uuid.Parse(creator)
	if err != nil {
		return nil, fmt.Errorf("invalid creator user id: %w", err)
	}

	// check for user exists
	u, err := s.userRepository.GetByID(ctx, creatorUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if u == nil || !u.IsActive {
		return nil, fmt.Errorf("creator not found or inactive")
	}

	g := group.NewGroup(name, creatorUUID, avatarURL)
	if err := s.groupRepository.Create(ctx, g); err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	// creator is owner
	ownerMember := group.NewMember(g.ID, creatorUUID, group.RoleOwner)
	if err := s.groupMemberRepository.AddMember(ctx, ownerMember); err != nil {
		return nil, fmt.Errorf("failed to add owner as group member: %w", err)
	}

	// optional members
	for _, member := range members {
		member = trimStringSpace(member)
		if member == "" {
			continue
		}

		memberUUID, err := uuid.Parse(member)
		if err != nil {
			continue
		}

		if memberUUID == creatorUUID {
			continue
		}

		other, err := s.userRepository.GetByID(ctx, memberUUID)
		if err != nil || other == nil || !other.IsActive {
			continue
		}

		m := group.NewMember(g.ID, memberUUID, group.RoleMember)
		_ = s.groupMemberRepository.AddMember(ctx, m) // if conflict - just update
	}

	resp := &GroupResponse{
		ID:        g.ID.String(),
		Name:      g.Name,
		AvatarURL: g.AvatarURL,
		Role:      string(group.RoleOwner),
	}

	membersList, _ := s.groupMemberRepository.List(ctx, g.ID)
	var ids []uuid.UUID
	for _, m := range membersList {
		if m.IsActive {
			ids = append(ids, m.UserID)
		}
	}
	s.hub.SetGroupMembers(g.ID, ids)

	if s.syncRepository != nil {
		if _, err := s.syncRepository.Append(ctx, creatorUUID, "group:created", map[string]string{
			"group_id": g.ID.String(),
			"name":     g.Name,
		}); err != nil {
			s.log.Warn("sync append failed", zap.Error(err))
		}
	}

	return resp, nil
}

// ListUserGroups returns all active groups of user
func (s *Service) ListUserGroups(ctx context.Context, userID string) ([]GroupResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	gs, err := s.groupRepository.ListByUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	resp := make([]GroupResponse, 0, len(gs))

	for _, g := range gs {
		role, err := s.groupMemberRepository.GetRole(ctx, g.ID, uid)
		if err != nil {
			continue // if no role, so skip :(
		}

		resp = append(resp, GroupResponse{
			ID:        g.ID.String(),
			Role:      string(role),
			Name:      g.Name,
			AvatarURL: g.AvatarURL,
		})
	}

	return resp, nil
}

func (s *Service) ListMembers(ctx context.Context, userID, groupID string) ([]GroupMemberResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	gid, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group id: %w", err)
	}

	isMember, err := s.groupMemberRepository.IsMember(ctx, gid, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to check group membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("user not a member of this group")
	}

	members, err := s.groupMemberRepository.List(ctx, gid)
	if err != nil {
		return nil, fmt.Errorf("failed to list group members: %w", err)
	}

	res := make([]GroupMemberResponse, 0, len(members))
	for _, m := range members {
		res = append(res, GroupMemberResponse{
			UserID:   m.UserID.String(),
			Role:     string(m.Role),
			IsActive: m.IsActive,
		})
	}

	return res, nil
}

// AddMember allows to add a member to group
func (s *Service) AddMember(ctx context.Context, actorUserID, groupID, targetUserID string) error {
	actorUUID, err := uuid.Parse(actorUserID)
	if err != nil {
		return fmt.Errorf("invalid actor user id: %w", err)
	}

	groupUUID, err := uuid.Parse(groupID)
	if err != nil {
		return fmt.Errorf("invalid group id: %w", err)
	}

	targetUUID, err := uuid.Parse(targetUserID)
	if err != nil {
		return fmt.Errorf("invalid target user id: %w", err)
	}

	role, err := s.groupMemberRepository.GetRole(ctx, groupUUID, actorUUID)
	if err != nil {
		return fmt.Errorf("failed to get actor role: %w", err)
	}
	if role != group.RoleOwner && role != group.RoleAdmin {
		return fmt.Errorf("only owner/admin cann add members")
	}

	if actorUUID == targetUUID {
		return fmt.Errorf("user (actor) is already a member")
	}

	isMember, err := s.groupMemberRepository.IsMember(ctx, groupUUID, targetUUID)
	if err != nil {
		return fmt.Errorf("failed to check target group membership: %w", err)
	}
	if isMember {
		return fmt.Errorf("user (target) is already a member")
	}

	u, err := s.userRepository.GetByID(ctx, targetUUID)
	if err != nil {
		return fmt.Errorf("failed to get target user: %w", err)
	}
	if u == nil || !u.IsActive {
		return fmt.Errorf("target user not found or inactive")
	}

	m := group.NewMember(groupUUID, targetUUID, group.RoleMember)
	if err := s.groupMemberRepository.AddMember(ctx, m); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	membersList, _ := s.groupMemberRepository.List(ctx, groupUUID)
	var ids []uuid.UUID
	for _, m := range membersList {
		if m.IsActive {
			ids = append(ids, m.UserID)
		}
	}
	s.hub.SetGroupMembers(groupUUID, ids)

	if s.syncRepository != nil {
		if _, err := s.syncRepository.Append(ctx, targetUUID, "group:member:add", map[string]string{
			"group_id": groupUUID.String(),
			"by":       actorUUID.String(),
			"role":     string(group.RoleMember),
		}); err != nil {
			s.log.Warn("sync append failed", zap.Error(err))
		}
	}

	return nil
}

// RemoveMember allows to remove user from group
func (s *Service) RemoveMember(ctx context.Context, actorUserID, groupID, targetUserID string) error {
	actorUUID, err := uuid.Parse(actorUserID)
	if err != nil {
		return fmt.Errorf("invalid actor user id: %w", err)
	}
	groupUUID, err := uuid.Parse(groupID)
	if err != nil {
		return fmt.Errorf("invalid group id: %w", err)
	}
	targetUUID, err := uuid.Parse(targetUserID)
	if err != nil {
		return fmt.Errorf("invalid target user id: %w", err)
	}

	role, err := s.groupMemberRepository.GetRole(ctx, groupUUID, actorUUID)
	if err != nil {
		return fmt.Errorf("failed to get actor role: %w", err)
	}
	if role != group.RoleOwner && role != group.RoleAdmin {
		return fmt.Errorf("only owner/admin cann remove members")
	}

	if actorUUID == targetUUID {
		return fmt.Errorf("cannot remove yourself")
	}

	isMember, err := s.groupMemberRepository.IsMember(ctx, groupUUID, targetUUID)
	if err != nil {
		return fmt.Errorf("failed to check target group membership: %w", err)
	}
	if !isMember {
		return fmt.Errorf("user (target) is not a member")
	}

	if err := s.groupMemberRepository.RemoveMember(ctx, groupUUID, targetUUID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	membersList, _ := s.groupMemberRepository.List(ctx, groupUUID)
	var ids []uuid.UUID
	for _, m := range membersList {
		if m.IsActive {
			ids = append(ids, m.UserID)
		}
	}
	s.hub.SetGroupMembers(groupUUID, ids)

	eventType := "group:member:remove"
	if actorUUID == targetUUID {
		eventType = "group:member:leave"
	}

	if s.syncRepository != nil {
		if _, err := s.syncRepository.Append(ctx, targetUUID, eventType, map[string]string{
			"group_id": groupUUID.String(),
			"by":       actorUUID.String(),
		}); err != nil {
			s.log.Warn("sync append failed", zap.Error(err))
		}
	}

	return nil
}

func (s *Service) GetActiveMemberIDs(ctx context.Context, actorID, groupID string) ([]uuid.UUID, error) {
	actorUID, err := uuid.Parse(actorID)
	if err != nil {
		return nil, fmt.Errorf("invalid actor id")
	}

	gid, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group id")
	}

	isMember, err := s.groupMemberRepository.IsMember(ctx, gid, actorUID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("not a member of group")
	}

	members, err := s.groupMemberRepository.List(ctx, gid)
	if err != nil {
		return nil, err
	}

	var ids []uuid.UUID
	for _, m := range members {
		if m.IsActive {
			ids = append(ids, m.UserID)
		}
	}

	return ids, nil
}

func (s *Service) EnsureMember(ctx context.Context, gid, uid string) error {
	userID, err := uuid.Parse(uid)
	if err != nil {
		return fmt.Errorf("invalid user id")
	}
	groupID, err := uuid.Parse(gid)
	if err != nil {
		return fmt.Errorf("invalid group id")
	}

	isMember, err := s.groupMemberRepository.IsMember(ctx, groupID, userID)
	if err != nil {
		return fmt.Errorf("failed to check target group membership: %w", err)
	}
	if !isMember {
		return fmt.Errorf("user is not a member of group")
	}

	return nil
}

// trimStringSpace remove spaces and other things
func trimStringSpace(s string) string {
	i := 0
	j := len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\n' || s[j-1] == '\r') {
		j--
	}
	return s[i:j]
}
