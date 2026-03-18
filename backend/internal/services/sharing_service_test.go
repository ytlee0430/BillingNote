package services

import (
	"billing-note/internal/models"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Sharing Repository ---

type mockSharingRepo struct {
	mock.Mock
}

func (m *mockSharingRepo) GetPairingCode(userID uint) (*models.UserPairingCode, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserPairingCode), args.Error(1)
}

func (m *mockSharingRepo) SavePairingCode(code *models.UserPairingCode) error {
	args := m.Called(code)
	return args.Error(0)
}

func (m *mockSharingRepo) FindByCode(code string) (*models.UserPairingCode, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserPairingCode), args.Error(1)
}

func (m *mockSharingRepo) CreateSharedAccess(access *models.SharedAccess) error {
	args := m.Called(access)
	return args.Error(0)
}

func (m *mockSharingRepo) ListSharedByOwner(ownerID uint) ([]models.SharedAccess, error) {
	args := m.Called(ownerID)
	return args.Get(0).([]models.SharedAccess), args.Error(1)
}

func (m *mockSharingRepo) ListSharedByViewer(viewerID uint) ([]models.SharedAccess, error) {
	args := m.Called(viewerID)
	return args.Get(0).([]models.SharedAccess), args.Error(1)
}

func (m *mockSharingRepo) DeleteSharedAccess(ownerID, viewerID uint) error {
	args := m.Called(ownerID, viewerID)
	return args.Error(0)
}

func (m *mockSharingRepo) HasAccess(ownerID, viewerID uint) (bool, error) {
	args := m.Called(ownerID, viewerID)
	return args.Bool(0), args.Error(1)
}

// --- Tests ---

func TestGetOrCreateCode_ExistingCode(t *testing.T) {
	repo := new(mockSharingRepo)
	svc := NewSharingService(repo)

	existing := &models.UserPairingCode{ID: 1, UserID: 1, Code: "AB12-CD34"}
	repo.On("GetPairingCode", uint(1)).Return(existing, nil)

	code, err := svc.GetOrCreateCode(1)

	assert.NoError(t, err)
	assert.Equal(t, "AB12-CD34", code.Code)
	repo.AssertExpectations(t)
}

func TestGetOrCreateCode_NewCode(t *testing.T) {
	repo := new(mockSharingRepo)
	svc := NewSharingService(repo)

	repo.On("GetPairingCode", uint(1)).Return(nil, nil)
	repo.On("SavePairingCode", mock.AnythingOfType("*models.UserPairingCode")).Return(nil)

	code, err := svc.GetOrCreateCode(1)

	assert.NoError(t, err)
	assert.NotEmpty(t, code.Code)
	assert.Len(t, code.Code, 9) // AB12-CD34 = 9 chars
	assert.Equal(t, '-', rune(code.Code[4]))
	repo.AssertExpectations(t)
}

func TestRegenerateCode(t *testing.T) {
	repo := new(mockSharingRepo)
	svc := NewSharingService(repo)

	existing := &models.UserPairingCode{ID: 1, UserID: 1, Code: "AB12-CD34"}
	repo.On("GetPairingCode", uint(1)).Return(existing, nil)
	repo.On("SavePairingCode", mock.AnythingOfType("*models.UserPairingCode")).Return(nil)

	code, err := svc.RegenerateCode(1)

	assert.NoError(t, err)
	assert.NotEmpty(t, code.Code)
	assert.Len(t, code.Code, 9)
	repo.AssertExpectations(t)
}

func TestPair_Success(t *testing.T) {
	repo := new(mockSharingRepo)
	svc := NewSharingService(repo)

	pairingCode := &models.UserPairingCode{ID: 1, UserID: 1, Code: "AB12-CD34"}
	repo.On("FindByCode", "AB12-CD34").Return(pairingCode, nil)
	repo.On("HasAccess", uint(1), uint(2)).Return(false, nil)
	repo.On("CreateSharedAccess", mock.AnythingOfType("*models.SharedAccess")).Return(nil)

	err := svc.Pair(2, "AB12-CD34")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestPair_InvalidCode(t *testing.T) {
	repo := new(mockSharingRepo)
	svc := NewSharingService(repo)

	repo.On("FindByCode", "XXXX-YYYY").Return(nil, nil)

	err := svc.Pair(2, "XXXX-YYYY")

	assert.Error(t, err)
	assert.Equal(t, "invalid pairing code", err.Error())
}

func TestPair_SelfPairing(t *testing.T) {
	repo := new(mockSharingRepo)
	svc := NewSharingService(repo)

	pairingCode := &models.UserPairingCode{ID: 1, UserID: 1, Code: "AB12-CD34"}
	repo.On("FindByCode", "AB12-CD34").Return(pairingCode, nil)

	err := svc.Pair(1, "AB12-CD34")

	assert.Error(t, err)
	assert.Equal(t, "cannot pair with yourself", err.Error())
}

func TestPair_AlreadyPaired(t *testing.T) {
	repo := new(mockSharingRepo)
	svc := NewSharingService(repo)

	pairingCode := &models.UserPairingCode{ID: 1, UserID: 1, Code: "AB12-CD34"}
	repo.On("FindByCode", "AB12-CD34").Return(pairingCode, nil)
	repo.On("HasAccess", uint(1), uint(2)).Return(true, nil)

	err := svc.Pair(2, "AB12-CD34")

	assert.Error(t, err)
	assert.Equal(t, "already paired with this user", err.Error())
}

func TestListViewers(t *testing.T) {
	repo := new(mockSharingRepo)
	svc := NewSharingService(repo)

	accesses := []models.SharedAccess{
		{ID: 1, OwnerID: 1, ViewerID: 2},
	}
	repo.On("ListSharedByOwner", uint(1)).Return(accesses, nil)

	result, err := svc.ListViewers(1)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, uint(2), result[0].ViewerID)
}

func TestRevoke(t *testing.T) {
	repo := new(mockSharingRepo)
	svc := NewSharingService(repo)

	repo.On("DeleteSharedAccess", uint(1), uint(2)).Return(nil)

	err := svc.Revoke(1, 2)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetOrCreateCode_RepoError(t *testing.T) {
	repo := new(mockSharingRepo)
	svc := NewSharingService(repo)

	repo.On("GetPairingCode", uint(1)).Return(nil, errors.New("db error"))

	_, err := svc.GetOrCreateCode(1)

	assert.Error(t, err)
}

func TestCodeFormat(t *testing.T) {
	repo := new(mockSharingRepo)
	svc := NewSharingService(repo)

	// Generate multiple codes and verify format
	for i := 0; i < 10; i++ {
		code, err := svc.generateCode()
		assert.NoError(t, err)
		assert.Len(t, code, 9)
		assert.Equal(t, byte('-'), code[4])

		// Verify letter positions (0,1,5,6) are uppercase letters
		for _, pos := range []int{0, 1, 5, 6} {
			assert.True(t, code[pos] >= 'A' && code[pos] <= 'Z', "position %d should be letter, got %c", pos, code[pos])
		}
		// Verify digit positions (2,3,7,8) are digits
		for _, pos := range []int{2, 3, 7, 8} {
			assert.True(t, code[pos] >= '0' && code[pos] <= '9', "position %d should be digit, got %c", pos, code[pos])
		}
	}
}
