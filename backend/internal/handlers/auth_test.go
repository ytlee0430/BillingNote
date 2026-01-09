package handlers

import (
	"billing-note/internal/services"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(req *services.RegisterRequest) (*services.AuthResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func (m *MockAuthService) Login(req *services.LoginRequest) (*services.AuthResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func setupAuthTest() (*gin.Engine, *MockAuthService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	router.POST("/auth/register", handler.Register)
	router.POST("/auth/login", handler.Login)
	router.GET("/auth/me", handler.Me)

	return router, mockService
}

func TestAuthHandler_Register_Success(t *testing.T) {
	router, mockService := setupAuthTest()

	registerReq := &services.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	mockResponse := &services.AuthResponse{
		Token: "mock-jwt-token",
		User:  nil, // Simplified for test
	}

	mockService.On("Register", mock.AnythingOfType("*services.RegisterRequest")).Return(mockResponse, nil)

	body, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response services.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "mock-jwt-token", response.Token)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Register_Failure_EmailExists(t *testing.T) {
	router, mockService := setupAuthTest()

	registerReq := &services.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	mockService.On("Register", mock.AnythingOfType("*services.RegisterRequest")).
		Return(nil, errors.New("email already registered"))

	body, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "email already registered", response["error"])

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Register_Failure_InvalidJSON(t *testing.T) {
	router, _ := setupAuthTest()

	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_Failure_MissingFields(t *testing.T) {
	router, _ := setupAuthTest()

	// Missing required email field
	invalidReq := map[string]string{
		"password": "password123",
	}

	body, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	router, mockService := setupAuthTest()

	loginReq := &services.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	mockResponse := &services.AuthResponse{
		Token: "mock-jwt-token",
		User:  nil,
	}

	mockService.On("Login", mock.AnythingOfType("*services.LoginRequest")).Return(mockResponse, nil)

	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response services.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "mock-jwt-token", response.Token)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_Failure_InvalidCredentials(t *testing.T) {
	router, mockService := setupAuthTest()

	loginReq := &services.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockService.On("Login", mock.AnythingOfType("*services.LoginRequest")).
		Return(nil, errors.New("invalid email or password"))

	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid email or password", response["error"])

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_Failure_InvalidJSON(t *testing.T) {
	router, _ := setupAuthTest()

	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Me_Success(t *testing.T) {
	router, _ := setupAuthTest()

	req, _ := http.NewRequest(http.MethodGet, "/auth/me", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Simulate authenticated user context
	c.Set("user_id", uint(1))
	c.Set("user_email", "test@example.com")

	handler := NewAuthHandler(nil)
	handler.Me(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), response["user_id"])
	assert.Equal(t, "test@example.com", response["email"])
}

func TestAuthHandler_Me_Failure_Unauthenticated(t *testing.T) {
	router, _ := setupAuthTest()

	req, _ := http.NewRequest(http.MethodGet, "/auth/me", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// No user context set - simulating unauthenticated request

	handler := NewAuthHandler(nil)
	handler.Me(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", response["error"])
}
