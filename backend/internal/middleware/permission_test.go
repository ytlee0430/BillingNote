package middleware

import (
	"billing-note/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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

func setupTestRouter(repo *mockSharingRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestViewAsMiddleware_NoViewAs(t *testing.T) {
	repo := new(mockSharingRepo)
	r := setupTestRouter(repo)

	r.GET("/test", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	}, ViewAsMiddleware(repo), func(c *gin.Context) {
		dataUserID, _ := GetDataUserID(c)
		readOnly := IsReadOnly(c)
		c.JSON(200, gin.H{"data_user_id": dataUserID, "read_only": readOnly})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"data_user_id":1`)
	assert.Contains(t, w.Body.String(), `"read_only":false`)
}

func TestViewAsMiddleware_ViewAsSelf(t *testing.T) {
	repo := new(mockSharingRepo)
	r := setupTestRouter(repo)

	r.GET("/test", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	}, ViewAsMiddleware(repo), func(c *gin.Context) {
		dataUserID, _ := GetDataUserID(c)
		readOnly := IsReadOnly(c)
		c.JSON(200, gin.H{"data_user_id": dataUserID, "read_only": readOnly})
	})

	req, _ := http.NewRequest("GET", "/test?view_as=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"data_user_id":1`)
	assert.Contains(t, w.Body.String(), `"read_only":false`)
}

func TestViewAsMiddleware_ViewAsWithAccess(t *testing.T) {
	repo := new(mockSharingRepo)
	r := setupTestRouter(repo)

	// User 2 (owner) granted access to User 1 (viewer)
	repo.On("HasAccess", uint(2), uint(1)).Return(true, nil)

	r.GET("/test", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	}, ViewAsMiddleware(repo), func(c *gin.Context) {
		dataUserID, _ := GetDataUserID(c)
		readOnly := IsReadOnly(c)
		c.JSON(200, gin.H{"data_user_id": dataUserID, "read_only": readOnly})
	})

	req, _ := http.NewRequest("GET", "/test?view_as=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"data_user_id":2`)
	assert.Contains(t, w.Body.String(), `"read_only":true`)
}

func TestViewAsMiddleware_ViewAsWithoutAccess(t *testing.T) {
	repo := new(mockSharingRepo)
	r := setupTestRouter(repo)

	// User 3 did NOT grant access to User 1
	repo.On("HasAccess", uint(3), uint(1)).Return(false, nil)

	r.GET("/test", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	}, ViewAsMiddleware(repo), func(c *gin.Context) {
		c.JSON(200, gin.H{"should": "not reach"})
	})

	req, _ := http.NewRequest("GET", "/test?view_as=3", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
	assert.Contains(t, w.Body.String(), "you do not have access")
}

func TestViewAsMiddleware_InvalidViewAs(t *testing.T) {
	repo := new(mockSharingRepo)
	r := setupTestRouter(repo)

	r.GET("/test", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	}, ViewAsMiddleware(repo), func(c *gin.Context) {
		c.JSON(200, gin.H{"should": "not reach"})
	})

	req, _ := http.NewRequest("GET", "/test?view_as=abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "invalid view_as")
}

func TestReadOnlyGuard_AllowsGetInReadOnly(t *testing.T) {
	r := gin.New()
	gin.SetMode(gin.TestMode)

	r.GET("/test", func(c *gin.Context) {
		c.Set("read_only", true)
		c.Next()
	}, ReadOnlyGuard(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestReadOnlyGuard_BlocksPostInReadOnly(t *testing.T) {
	r := gin.New()
	gin.SetMode(gin.TestMode)

	r.POST("/test", func(c *gin.Context) {
		c.Set("read_only", true)
		c.Next()
	}, ReadOnlyGuard(), func(c *gin.Context) {
		c.JSON(200, gin.H{"should": "not reach"})
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
	assert.Contains(t, w.Body.String(), "read-only mode")
}

func TestReadOnlyGuard_BlocksPutInReadOnly(t *testing.T) {
	r := gin.New()
	gin.SetMode(gin.TestMode)

	r.PUT("/test", func(c *gin.Context) {
		c.Set("read_only", true)
		c.Next()
	}, ReadOnlyGuard(), func(c *gin.Context) {
		c.JSON(200, gin.H{"should": "not reach"})
	})

	req, _ := http.NewRequest("PUT", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
}

func TestReadOnlyGuard_BlocksDeleteInReadOnly(t *testing.T) {
	r := gin.New()
	gin.SetMode(gin.TestMode)

	r.DELETE("/test", func(c *gin.Context) {
		c.Set("read_only", true)
		c.Next()
	}, ReadOnlyGuard(), func(c *gin.Context) {
		c.JSON(200, gin.H{"should": "not reach"})
	})

	req, _ := http.NewRequest("DELETE", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
}

func TestReadOnlyGuard_AllowsWriteWhenNotReadOnly(t *testing.T) {
	r := gin.New()
	gin.SetMode(gin.TestMode)

	r.POST("/test", func(c *gin.Context) {
		c.Set("read_only", false)
		c.Next()
	}, ReadOnlyGuard(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestViewAsMiddleware_Unauthenticated(t *testing.T) {
	repo := new(mockSharingRepo)
	r := setupTestRouter(repo)

	// No user_id set in context
	r.GET("/test", ViewAsMiddleware(repo), func(c *gin.Context) {
		c.JSON(200, gin.H{"should": "not reach"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}
