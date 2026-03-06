package handlers

import (
	"billing-note/internal/models"
	"billing-note/internal/services"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mock Gmail Repository for handler tests ---

type mockGmailRepoHandler struct {
	mock.Mock
}

func (m *mockGmailRepoHandler) SaveToken(token *models.GmailToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *mockGmailRepoHandler) GetToken(userID uint) (*models.GmailToken, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GmailToken), args.Error(1)
}

func (m *mockGmailRepoHandler) DeleteToken(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *mockGmailRepoHandler) GetScanRule(userID uint) (*models.GmailScanRule, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GmailScanRule), args.Error(1)
}

func (m *mockGmailRepoHandler) SaveScanRule(rule *models.GmailScanRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *mockGmailRepoHandler) CreateScanHistory(history *models.GmailScanHistory) error {
	args := m.Called(history)
	return args.Error(0)
}

func (m *mockGmailRepoHandler) ListScanHistory(userID uint, limit int) ([]models.GmailScanHistory, error) {
	args := m.Called(userID, limit)
	return args.Get(0).([]models.GmailScanHistory), args.Error(1)
}

// --- Setup ---

const handlerTestEncKey = "test-encryption-key-32-bytes-ok!"
const handlerTestJWTSecret = "test-jwt-secret"

func setupGmailTestRouter() (*gin.Engine, *mockGmailRepoHandler) {
	gin.SetMode(gin.TestMode)

	repo := new(mockGmailRepoHandler)

	// Default mock behaviors
	repo.On("GetToken", mock.Anything).Return(nil, gorm.ErrRecordNotFound).Maybe()
	repo.On("GetScanRule", mock.Anything).Return(nil, gorm.ErrRecordNotFound).Maybe()
	repo.On("SaveScanRule", mock.Anything).Return(nil).Maybe()
	repo.On("SaveToken", mock.Anything).Return(nil).Maybe()
	repo.On("DeleteToken", mock.Anything).Return(nil).Maybe()

	svc, _ := services.NewGmailService(
		repo,
		handlerTestEncKey,
		"test-client-id",
		"test-client-secret",
		"http://localhost/callback",
		handlerTestJWTSecret,
	)
	scanSvc := services.NewGmailScanService(svc, nil, repo, "/tmp/test-uploads")
	handler := NewGmailHandler(svc, scanSvc)

	r := gin.New()
	api := r.Group("/api")
	api.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("user_email", "test@example.com")
		c.Set("request_id", "test-request-id")
		c.Next()
	})

	api.GET("/gmail/auth", handler.GetAuthURL)
	api.POST("/gmail/callback", handler.HandleCallback)
	api.GET("/gmail/status", handler.GetStatus)
	api.PUT("/gmail/settings", handler.UpdateSettings)
	api.DELETE("/gmail/disconnect", handler.Disconnect)

	return r, repo
}

// --- Tests ---

func TestGetAuthURL_Integration(t *testing.T) {
	r, _ := setupGmailTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/gmail/auth", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["url"], "accounts.google.com")
	assert.Contains(t, resp["url"], "test-client-id")
}

func TestHandleCallback_MissingFields(t *testing.T) {
	r, _ := setupGmailTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/gmail/callback",
		strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCallback_InvalidState(t *testing.T) {
	r, _ := setupGmailTestRouter()

	w := httptest.NewRecorder()
	body := `{"code":"test-code","state":"invalid-state"}`
	req, _ := http.NewRequest("POST", "/api/gmail/callback",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetStatus_NotConnected_Integration(t *testing.T) {
	r, _ := setupGmailTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/gmail/status", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["connected"])
}

func TestDisconnect_NotConnected_Integration(t *testing.T) {
	r, _ := setupGmailTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/gmail/disconnect", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateSettings_Integration(t *testing.T) {
	r, _ := setupGmailTestRouter()

	w := httptest.NewRecorder()
	body := `{"enabled":true,"sender_keywords":["test"]}`
	req, _ := http.NewRequest("PUT", "/api/gmail/settings",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Gmail settings updated", resp["message"])
}

func TestUpdateSettings_InvalidBody(t *testing.T) {
	r, _ := setupGmailTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/gmail/settings",
		strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
