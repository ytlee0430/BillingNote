package services

import (
	"billing-note/internal/models"
	"billing-note/internal/repository"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

type SharingService struct {
	repo repository.SharingRepository
}

func NewSharingService(repo repository.SharingRepository) *SharingService {
	return &SharingService{repo: repo}
}

// GetOrCreateCode returns the user's existing pairing code or generates a new one.
func (s *SharingService) GetOrCreateCode(userID uint) (*models.UserPairingCode, error) {
	existing, err := s.repo.GetPairingCode(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pairing code: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	code, err := s.generateCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code: %w", err)
	}

	pairingCode := &models.UserPairingCode{
		UserID: userID,
		Code:   code,
	}
	if err := s.repo.SavePairingCode(pairingCode); err != nil {
		return nil, fmt.Errorf("failed to save pairing code: %w", err)
	}

	return pairingCode, nil
}

// RegenerateCode creates a new pairing code for the user.
func (s *SharingService) RegenerateCode(userID uint) (*models.UserPairingCode, error) {
	existing, err := s.repo.GetPairingCode(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pairing code: %w", err)
	}

	code, err := s.generateCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code: %w", err)
	}

	if existing != nil {
		existing.Code = code
		if err := s.repo.SavePairingCode(existing); err != nil {
			return nil, fmt.Errorf("failed to save pairing code: %w", err)
		}
		return existing, nil
	}

	pairingCode := &models.UserPairingCode{
		UserID: userID,
		Code:   code,
	}
	if err := s.repo.SavePairingCode(pairingCode); err != nil {
		return nil, fmt.Errorf("failed to save pairing code: %w", err)
	}

	return pairingCode, nil
}

// Pair links the viewer to the owner via pairing code.
func (s *SharingService) Pair(viewerID uint, code string) error {
	pairingCode, err := s.repo.FindByCode(code)
	if err != nil {
		return fmt.Errorf("failed to find pairing code: %w", err)
	}
	if pairingCode == nil {
		return errors.New("invalid pairing code")
	}

	if pairingCode.UserID == viewerID {
		return errors.New("cannot pair with yourself")
	}

	// Check if already paired
	hasAccess, err := s.repo.HasAccess(pairingCode.UserID, viewerID)
	if err != nil {
		return fmt.Errorf("failed to check access: %w", err)
	}
	if hasAccess {
		return errors.New("already paired with this user")
	}

	access := &models.SharedAccess{
		OwnerID:  pairingCode.UserID,
		ViewerID: viewerID,
	}
	if err := s.repo.CreateSharedAccess(access); err != nil {
		return fmt.Errorf("failed to create shared access: %w", err)
	}

	return nil
}

// ListViewers returns users who have view access to the owner's data.
func (s *SharingService) ListViewers(ownerID uint) ([]models.SharedAccess, error) {
	return s.repo.ListSharedByOwner(ownerID)
}

// ListOwners returns users whose data the viewer can see.
func (s *SharingService) ListOwners(viewerID uint) ([]models.SharedAccess, error) {
	return s.repo.ListSharedByViewer(viewerID)
}

// Revoke removes a viewer's access to the owner's data.
func (s *SharingService) Revoke(ownerID, viewerID uint) error {
	return s.repo.DeleteSharedAccess(ownerID, viewerID)
}

// generateCode generates a code in AB12-CD34 format.
func (s *SharingService) generateCode() (string, error) {
	const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ" // exclude I, O to avoid confusion
	const digits = "0123456789"

	pick := func(charset string) (byte, error) {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return 0, err
		}
		return charset[n.Int64()], nil
	}

	// AB12-CD34 pattern: letter letter digit digit - letter letter digit digit
	parts := []string{letters, letters, digits, digits, letters, letters, digits, digits}
	result := make([]byte, 9) // 8 chars + 1 dash
	idx := 0
	for i, charset := range parts {
		if i == 4 {
			result[idx] = '-'
			idx++
		}
		ch, err := pick(charset)
		if err != nil {
			return "", err
		}
		result[idx] = ch
		idx++
	}

	return string(result), nil
}
