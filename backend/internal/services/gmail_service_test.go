package services

import (
	"billing-note/internal/models"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mock Gmail Repository ---

type mockGmailRepo struct {
	mock.Mock
}

func (m *mockGmailRepo) SaveToken(token *models.GmailToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *mockGmailRepo) GetToken(userID uint) (*models.GmailToken, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GmailToken), args.Error(1)
}

func (m *mockGmailRepo) DeleteToken(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *mockGmailRepo) GetScanRule(userID uint) (*models.GmailScanRule, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GmailScanRule), args.Error(1)
}

func (m *mockGmailRepo) SaveScanRule(rule *models.GmailScanRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *mockGmailRepo) CreateScanHistory(history *models.GmailScanHistory) error {
	args := m.Called(history)
	return args.Error(0)
}

func (m *mockGmailRepo) ListScanHistory(userID uint, limit int) ([]models.GmailScanHistory, error) {
	args := m.Called(userID, limit)
	return args.Get(0).([]models.GmailScanHistory), args.Error(1)
}

// --- Helpers ---

const testEncryptionKey = "test-encryption-key-32-bytes-ok!"
const testStateSecret = "test-state-secret"

func newTestGmailService(repo *mockGmailRepo) *GmailService {
	svc, _ := NewGmailService(
		repo,
		testEncryptionKey,
		"test-client-id",
		"test-client-secret",
		"http://localhost/callback",
		testStateSecret,
	)
	return svc
}

func makeExpiredState(userID uint, secret string) string {
	oldTimestamp := time.Now().Unix() - 700 // > 10 minutes ago
	data := fmt.Sprintf("%d:%d", userID, oldTimestamp)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%d:%d:%s", userID, oldTimestamp, sig)
}

// --- Tests ---

func TestNewGmailService_Success(t *testing.T) {
	repo := new(mockGmailRepo)
	svc, err := NewGmailService(
		repo,
		testEncryptionKey,
		"test-client-id",
		"test-client-secret",
		"http://localhost/callback",
		testStateSecret,
	)
	assert.NoError(t, err)
	assert.NotNil(t, svc)
	assert.Equal(t, "test-client-id", svc.oauthConfig.ClientID)
	assert.Equal(t, "test-client-secret", svc.oauthConfig.ClientSecret)
	assert.Equal(t, "http://localhost/callback", svc.oauthConfig.RedirectURL)
	assert.Len(t, svc.oauthConfig.Scopes, 2)
}

func TestGetAuthURL(t *testing.T) {
	repo := new(mockGmailRepo)
	svc := newTestGmailService(repo)

	url, err := svc.GetAuthURL(1)
	assert.NoError(t, err)
	assert.Contains(t, url, "accounts.google.com")
	assert.Contains(t, url, "test-client-id")
	assert.Contains(t, url, "state=")
}

func TestGetStatus_NotConnected(t *testing.T) {
	repo := new(mockGmailRepo)
	repo.On("GetToken", uint(1)).Return(nil, gorm.ErrRecordNotFound)

	svc := newTestGmailService(repo)

	status, err := svc.GetStatus(1)
	assert.NoError(t, err)
	assert.False(t, status.Connected)
	repo.AssertExpectations(t)
}

func TestGetStatus_Connected(t *testing.T) {
	repo := new(mockGmailRepo)
	svc := newTestGmailService(repo)

	accessEnc, _ := svc.crypto.Encrypt("test-access-token")
	refreshEnc, _ := svc.crypto.Encrypt("test-refresh-token")
	now := time.Now()

	token := &models.GmailToken{
		UserID:                1,
		AccessTokenEncrypted:  accessEnc,
		RefreshTokenEncrypted: refreshEnc,
		TokenExpiry:           &now,
		Scopes:                "gmail.readonly gmail.metadata",
		CreatedAt:             now,
	}
	repo.On("GetToken", uint(1)).Return(token, nil)
	repo.On("GetScanRule", uint(1)).Return(nil, gorm.ErrRecordNotFound)

	status, err := svc.GetStatus(1)
	assert.NoError(t, err)
	assert.True(t, status.Connected)
	assert.Equal(t, "gmail.readonly gmail.metadata", status.Scopes)
	assert.NotNil(t, status.ConnectedAt)
	repo.AssertExpectations(t)
}

func TestDisconnect_NotConnected(t *testing.T) {
	repo := new(mockGmailRepo)
	repo.On("GetToken", uint(1)).Return(nil, gorm.ErrRecordNotFound)

	svc := newTestGmailService(repo)

	err := svc.Disconnect(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	repo.AssertExpectations(t)
}

func TestDisconnect_Success(t *testing.T) {
	repo := new(mockGmailRepo)
	svc := newTestGmailService(repo)

	accessEnc, _ := svc.crypto.Encrypt("test-access-token")
	refreshEnc, _ := svc.crypto.Encrypt("test-refresh-token")
	now := time.Now()

	token := &models.GmailToken{
		UserID:                1,
		AccessTokenEncrypted:  accessEnc,
		RefreshTokenEncrypted: refreshEnc,
		TokenExpiry:           &now,
	}
	repo.On("GetToken", uint(1)).Return(token, nil)
	repo.On("DeleteToken", uint(1)).Return(nil)

	err := svc.Disconnect(1)
	assert.NoError(t, err)
	repo.AssertCalled(t, "DeleteToken", uint(1))
	repo.AssertExpectations(t)
}

func TestUpdateSettings_NewRule(t *testing.T) {
	repo := new(mockGmailRepo)
	repo.On("GetScanRule", uint(1)).Return(nil, gorm.ErrRecordNotFound)
	repo.On("SaveScanRule", mock.AnythingOfType("*models.GmailScanRule")).Return(nil)

	svc := newTestGmailService(repo)

	enabled := true
	input := models.GmailSettingsInput{
		Enabled:         &enabled,
		SenderKeywords:  []string{"test"},
		SubjectKeywords: []string{"bill"},
	}

	err := svc.UpdateSettings(1, input)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpdateSettings_ExistingRule(t *testing.T) {
	repo := new(mockGmailRepo)

	existingRule := &models.GmailScanRule{
		UserID:            1,
		Enabled:           false,
		SenderKeywords:    []string{"old"},
		SubjectKeywords:   []string{"old"},
		RequireAttachment: true,
	}
	repo.On("GetScanRule", uint(1)).Return(existingRule, nil)
	repo.On("SaveScanRule", mock.AnythingOfType("*models.GmailScanRule")).Return(nil)

	svc := newTestGmailService(repo)

	enabled := true
	input := models.GmailSettingsInput{
		Enabled:        &enabled,
		SenderKeywords: []string{"new-keyword"},
	}

	err := svc.UpdateSettings(1, input)
	assert.NoError(t, err)

	savedRule := repo.Calls[1].Arguments.Get(0).(*models.GmailScanRule)
	assert.True(t, savedRule.Enabled)
	assert.Equal(t, []string{"new-keyword"}, []string(savedRule.SenderKeywords))
	assert.Equal(t, []string{"old"}, []string(savedRule.SubjectKeywords))
	repo.AssertExpectations(t)
}

func TestGetSettings_Default(t *testing.T) {
	repo := new(mockGmailRepo)
	repo.On("GetScanRule", uint(1)).Return(nil, gorm.ErrRecordNotFound)

	svc := newTestGmailService(repo)

	rule, err := svc.GetSettings(1)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), rule.UserID)
	assert.False(t, rule.Enabled)
	assert.Contains(t, []string(rule.SenderKeywords), "credit")
	assert.Contains(t, []string(rule.SubjectKeywords), "帳單")
	repo.AssertExpectations(t)
}

func TestGenerateAndVerifyState(t *testing.T) {
	repo := new(mockGmailRepo)
	svc := newTestGmailService(repo)

	state := svc.generateState(42)
	assert.NotEmpty(t, state)

	// Should verify correctly for the same user
	assert.True(t, svc.verifyState(state, 42))

	// Should fail for a different user
	assert.False(t, svc.verifyState(state, 99))

	// Should fail for invalid state
	assert.False(t, svc.verifyState("invalid", 42))
	assert.False(t, svc.verifyState("", 42))
}

func TestVerifyState_Expired(t *testing.T) {
	repo := new(mockGmailRepo)
	svc := newTestGmailService(repo)

	expiredState := makeExpiredState(42, testStateSecret)
	assert.False(t, svc.verifyState(expiredState, 42))

	// Valid state should work
	validState := svc.generateState(42)
	assert.True(t, svc.verifyState(validState, 42))
}

func TestVerifyState_WrongSignature(t *testing.T) {
	repo := new(mockGmailRepo)
	svc := newTestGmailService(repo)

	fakeState := fmt.Sprintf("%d:%d:%s", 42, time.Now().Unix(), "fakesignature")
	assert.False(t, svc.verifyState(fakeState, 42))
}

func TestHandleCallback_InvalidState(t *testing.T) {
	repo := new(mockGmailRepo)
	svc := newTestGmailService(repo)

	err := svc.HandleCallback(1, "some-code", "invalid-state")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid OAuth state")
}
