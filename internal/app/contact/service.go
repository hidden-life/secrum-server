package contact

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/contact"
	"github.com/hidden-life/secrum-server/internal/ports"
)

type Service struct {
	userRepository    ports.UserRepository
	contactRepository ports.ContactRepository
}

type AddContactRequest struct {
	PhoneHash string `json:"phone_hash"`
}

type SyncContactRequest struct {
	PhoneHashes []string `json:"phone_hashes"`
}

type Profile struct {
	UserID            string  `json:"user_id"`
	DisplayName       *string `json:"display_name"`
	AvatarURL         *string `json:"avatar_url"`
	StatusMessage     *string `json:"status_message"`
	SafetyFingerprint *string `json:"safety_fingerprint"`
}

func NewService(userRepo ports.UserRepository, contactRepo ports.ContactRepository) *Service {
	return &Service{userRepository: userRepo, contactRepository: contactRepo}
}

// AddContact allows to add new contact
func (s *Service) AddContact(ctx context.Context, owner uuid.UUID, phoneHash string) error {
	// search a user by phone hash
	u, err := s.userRepository.GetByPhoneHash(ctx, phoneHash)
	if err != nil {
		return err
	}

	if u == nil {
		return fmt.Errorf("user not found")
	}

	c := contact.New(owner, u.ID)

	return s.contactRepository.Add(ctx, c)
}

func (s *Service) List(ctx context.Context, owner uuid.UUID) ([]*Profile, error) {
	contacts, err := s.contactRepository.List(ctx, owner)
	if err != nil {
		return nil, err
	}

	var result []*Profile
	for _, c := range contacts {
		u, err := s.userRepository.GetByID(ctx, c.ContactUserID)
		if err != nil || u == nil {
			continue
		}
		result = append(result, &Profile{
			UserID:            u.ID.String(),
			DisplayName:       u.DisplayName,
			AvatarURL:         u.AvatarURL,
			StatusMessage:     u.StatusMessage,
			SafetyFingerprint: u.SafetyFingerprint,
		})
	}

	return result, nil
}

func (s *Service) Sync(ctx context.Context, owner uuid.UUID, hashes []string) error {
	for _, h := range hashes {
		u, err := s.userRepository.GetByPhoneHash(ctx, h)
		if err != nil || u == nil {
			continue
		}

		if u.ID == owner {
			continue
		}

		_ = s.contactRepository.Add(ctx, contact.New(owner, u.ID))
	}

	return nil
}

func (s *Service) Remove(ctx context.Context, owner, target uuid.UUID) error {
	return s.contactRepository.Remove(ctx, owner, target)
}
