package integration

import (
	"billing-note/internal/handlers"
	"billing-note/internal/middleware"
	"billing-note/internal/models"
	"billing-note/internal/repository"
	"billing-note/internal/services"
	"billing-note/pkg/config"
	"billing-note/pkg/database"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupAuthTestServer(t *testing.T) (*gin.Engine, *gorm.DB, func()) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	gin.SetMode(gin.TestMode)

	// Setup test database
	os.Setenv("DB_NAME", "billing_note_test")
	cfg, err := config.Load()
	if err != nil {
		t.Skip("Skipping integration test: unable to load config")
		return nil, nil, nil
	}

	db, err := database.Connect(cfg.Database.DSN())
	if err != nil {
		t.Skip("Skipping integration test: unable to connect to database")
		return nil, nil, nil
	}

	// Setup repositories and services
	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.Expiry)

	// Setup handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Setup router
	router := gin.New()
	router.POST("/api/auth/register", authHandler.Register)
	router.POST("/api/auth/login", authHandler.Login)
	router.GET("/api/auth/me", middleware.AuthMiddleware(cfg.JWT.Secret), authHandler.Me)

	// Cleanup function
	cleanup := func() {
		// Clean up test data
		db.Exec("DELETE FROM users WHERE email LIKE '%@test.com'")
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return router, db, cleanup
}

func TestAuthAPI_RegisterFlow(t *testing.T) {
	router, db, cleanup := setupAuthTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	// Test successful registration
	registerReq := map[string]string{
		"email":    fmt.Sprintf("newuser_%d@test.com", time.Now().Unix()),
		"password": "password123",
		"name":     "Test User",
	}

	body, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response["token"])
	assert.NotNil(t, response["user"])

	// Verify user was created in database
	var user models.User
	err = db.Where("email = ?", registerReq["email"]).First(&user).Error
	assert.NoError(t, err)
	assert.Equal(t, registerReq["email"], user.Email)
	assert.Equal(t, registerReq["name"], user.Name)
}

func TestAuthAPI_RegisterDuplicateEmail(t *testing.T) {
	router, db, cleanup := setupAuthTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	email := fmt.Sprintf("duplicate_%d@test.com", time.Now().Unix())

	// Create first user
	user := &models.User{
		Email:        email,
		Name:         "First User",
		PasswordHash: "hashedpassword",
	}
	db.Create(user)

	// Try to register with same email
	registerReq := map[string]string{
		"email":    email,
		"password": "password123",
		"name":     "Second User",
	}

	body, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "already registered")
}

func TestAuthAPI_LoginFlow(t *testing.T) {
	router, db, cleanup := setupAuthTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	email := fmt.Sprintf("loginuser_%d@test.com", time.Now().Unix())
	password := "password123"

	// First register a user
	user := &models.User{
		Email:        email,
		Name:         "Login Test User",
		PasswordHash: password, // Will be hashed by BeforeCreate hook
	}
	db.Create(user)

	// Now try to login
	loginReq := map[string]string{
		"email":    email,
		"password": password,
	}

	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response["token"])
	assert.NotNil(t, response["user"])
}

func TestAuthAPI_LoginInvalidCredentials(t *testing.T) {
	router, _, cleanup := setupAuthTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	loginReq := map[string]string{
		"email":    "nonexistent@test.com",
		"password": "wrongpassword",
	}

	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid")
}

func TestAuthAPI_MeEndpoint(t *testing.T) {
	router, db, cleanup := setupAuthTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	email := fmt.Sprintf("metest_%d@test.com", time.Now().Unix())

	// Create a user
	user := &models.User{
		Email:        email,
		Name:         "Me Test User",
		PasswordHash: "password123",
	}
	db.Create(user)

	// Login to get token
	loginReq := map[string]string{
		"email":    email,
		"password": "password123",
	}

	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	token := loginResponse["token"].(string)

	// Now call /me endpoint with token
	req, _ = http.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var meResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &meResponse)
	assert.NoError(t, err)
	assert.NotNil(t, meResponse["user_id"])
	assert.Equal(t, email, meResponse["email"])
}

func TestAuthAPI_MeEndpointUnauthorized(t *testing.T) {
	router, _, cleanup := setupAuthTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	// Call /me endpoint without token
	req, _ := http.NewRequest(http.MethodGet, "/api/auth/me", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthAPI_FullRegistrationAndLoginFlow(t *testing.T) {
	router, _, cleanup := setupAuthTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	email := fmt.Sprintf("fullflow_%d@test.com", time.Now().Unix())
	password := "password123"

	// Step 1: Register
	registerReq := map[string]string{
		"email":    email,
		"password": password,
		"name":     "Full Flow User",
	}

	body, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var registerResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &registerResponse)
	registerToken := registerResponse["token"].(string)

	// Step 2: Login
	loginReq := map[string]string{
		"email":    email,
		"password": password,
	}

	body, _ = json.Marshal(loginReq)
	req, _ = http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	loginToken := loginResponse["token"].(string)

	// Step 3: Access protected endpoint with register token
	req, _ = http.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+registerToken)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Step 4: Access protected endpoint with login token
	req, _ = http.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+loginToken)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
