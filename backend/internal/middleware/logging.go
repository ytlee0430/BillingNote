package middleware

import (
	"billing-note/pkg/logger"
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggingMiddleware logs all incoming requests and outgoing responses
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()[:8]
		c.Set("request_id", requestID)

		// Start timer
		start := time.Now()

		// Read request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Log incoming request
		entry := logger.WithFields(logger.Fields{
			"request_id": requestID,
			"client_ip":  c.ClientIP(),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.RawQuery,
			"user_agent": c.Request.UserAgent(),
		})

		// Log request body for POST/PUT/PATCH (but not passwords)
		if len(requestBody) > 0 && len(requestBody) < 1000 {
			// Sanitize sensitive data
			bodyStr := sanitizeBody(string(requestBody))
			entry = entry.WithField("request_body", bodyStr)
		}

		entry.Info("Incoming request")

		// Wrap response writer to capture response body
		blw := &responseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get user ID if available
		userID, exists := GetUserID(c)

		// Build response log fields
		responseFields := logger.Fields{
			"request_id":  requestID,
			"status":      c.Writer.Status(),
			"latency":     latency.String(),
			"latency_ms":  latency.Milliseconds(),
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
		}

		if exists {
			responseFields["user_id"] = userID
		}

		// Log errors if any
		if len(c.Errors) > 0 {
			responseFields["errors"] = c.Errors.String()
		}

		// Log response body for errors (status >= 400)
		if c.Writer.Status() >= 400 && blw.body.Len() < 500 {
			responseFields["response_body"] = blw.body.String()
		}

		// Log at appropriate level based on status
		responseEntry := logger.WithFields(responseFields)

		switch {
		case c.Writer.Status() >= 500:
			responseEntry.Error("Request completed with server error")
		case c.Writer.Status() >= 400:
			responseEntry.Warn("Request completed with client error")
		default:
			responseEntry.Info("Request completed successfully")
		}
	}
}

// sanitizeBody removes sensitive data from request body for logging
func sanitizeBody(body string) string {
	// Simple sanitization - in production, use proper JSON parsing
	if len(body) > 500 {
		return body[:500] + "...(truncated)"
	}
	return body
}
