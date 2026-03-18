package middleware

import (
	"billing-note/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ViewAsMiddleware checks the view_as query parameter and validates access.
// If view_as is set, the data_user_id context is set to the target user.
// Otherwise, data_user_id defaults to the authenticated user.
// Also sets read_only=true when viewing another user's data.
func ViewAsMiddleware(sharingRepo repository.SharingRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		// Default: user views their own data
		c.Set("data_user_id", userID)
		c.Set("read_only", false)

		viewAsStr := c.Query("view_as")
		if viewAsStr != "" {
			viewAsID, err := strconv.ParseUint(viewAsStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid view_as parameter"})
				c.Abort()
				return
			}

			targetUserID := uint(viewAsID)

			// Cannot view_as yourself
			if targetUserID == userID {
				c.Set("data_user_id", userID)
				c.Set("read_only", false)
				c.Next()
				return
			}

			// Check if the authenticated user has access to the target user's data
			hasAccess, err := sharingRepo.HasAccess(targetUserID, userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check access"})
				c.Abort()
				return
			}

			if !hasAccess {
				c.JSON(http.StatusForbidden, gin.H{"error": "you do not have access to this user's data"})
				c.Abort()
				return
			}

			c.Set("data_user_id", targetUserID)
			c.Set("read_only", true)
		}

		c.Next()
	}
}

// ReadOnlyGuard rejects write operations when in read-only mode.
func ReadOnlyGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		readOnly, exists := c.Get("read_only")
		if exists && readOnly.(bool) {
			method := c.Request.Method
			if method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete || method == http.MethodPatch {
				c.JSON(http.StatusForbidden, gin.H{"error": "read-only mode: write operations are not allowed when viewing another user's data"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// GetDataUserID returns the effective user ID for data access.
func GetDataUserID(c *gin.Context) (uint, bool) {
	id, exists := c.Get("data_user_id")
	if !exists {
		return GetUserID(c)
	}
	return id.(uint), true
}

// IsReadOnly returns whether the current context is in read-only mode.
func IsReadOnly(c *gin.Context) bool {
	readOnly, exists := c.Get("read_only")
	if !exists {
		return false
	}
	return readOnly.(bool)
}
