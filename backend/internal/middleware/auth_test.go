package middleware

import (
	"billing-note/pkg/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupAuthMiddlewareTest() (*gin.Engine, string) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	jwtSecret := "test-secret-key"

	// Protected route
	router.GET("/protected", AuthMiddleware(jwtSecret), func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in context"})
			return
		}
		userEmail, _ := c.Get("user_email")
		c.JSON(http.StatusOK, gin.H{
			"message":    "success",
			"user_id":    userID,
			"user_email": userEmail,
		})
	})

	return router, jwtSecret
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	router, jwtSecret := setupAuthMiddlewareTest()

	// Generate a valid token
	token, err := utils.GenerateToken(1, "test@example.com", jwtSecret, 1*time.Hour)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response contains user info
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "test@example.com")
}

func TestAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	router, _ := setupAuthMiddlewareTest()

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authorization header required")
}

func TestAuthMiddleware_InvalidAuthorizationFormat_MissingBearer(t *testing.T) {
	router, jwtSecret := setupAuthMiddlewareTest()

	token, _ := utils.GenerateToken(1, "test@example.com", jwtSecret, 1*time.Hour)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", token) // Missing "Bearer " prefix

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid authorization header format")
}

func TestAuthMiddleware_InvalidAuthorizationFormat_WrongPrefix(t *testing.T) {
	router, jwtSecret := setupAuthMiddlewareTest()

	token, _ := utils.GenerateToken(1, "test@example.com", jwtSecret, 1*time.Hour)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic "+token) // Wrong prefix

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid authorization header format")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router, _ := setupAuthMiddlewareTest()

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	router, jwtSecret := setupAuthMiddlewareTest()

	// Generate a token that expires immediately
	token, err := utils.GenerateToken(1, "test@example.com", jwtSecret, -1*time.Hour)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

func TestAuthMiddleware_TokenWithWrongSecret(t *testing.T) {
	router, _ := setupAuthMiddlewareTest()

	// Generate token with different secret
	wrongSecret := "wrong-secret-key"
	token, err := utils.GenerateToken(1, "test@example.com", wrongSecret, 1*time.Hour)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

func TestAuthMiddleware_EmptyBearerToken(t *testing.T) {
	router, _ := setupAuthMiddlewareTest()

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetUserID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("user_id", uint(123))

	userID, exists := GetUserID(c)

	assert.True(t, exists)
	assert.Equal(t, uint(123), userID)
}

func TestGetUserID_NotExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Don't set user_id

	userID, exists := GetUserID(c)

	assert.False(t, exists)
	assert.Equal(t, uint(0), userID)
}

func TestAuthMiddleware_ContextValuesSetCorrectly(t *testing.T) {
	router, jwtSecret := setupAuthMiddlewareTest()

	testUserID := uint(42)
	testEmail := "contexttest@example.com"

	token, err := utils.GenerateToken(testUserID, testEmail, jwtSecret, 1*time.Hour)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the exact user_id is returned
	assert.Contains(t, w.Body.String(), "42")
	assert.Contains(t, w.Body.String(), testEmail)
}
