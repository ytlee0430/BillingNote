package logger

import (
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Init initializes the global logger
func Init(level string, jsonFormat bool) {
	log = logrus.New()

	// Set log level
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	// Set output
	log.SetOutput(os.Stdout)

	// Set format
	if jsonFormat {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}
}

// GetLogger returns the global logger
func GetLogger() *logrus.Logger {
	if log == nil {
		Init("info", false)
	}
	return log
}

// Fields type for structured logging
type Fields = logrus.Fields

// WithFields returns a new entry with fields
func WithFields(fields Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}

// WithError returns a new entry with error
func WithError(err error) *logrus.Entry {
	return GetLogger().WithError(err)
}

// WithField returns a new entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}

// Debug logs debug level message
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Debugf logs debug level formatted message
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// Info logs info level message
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof logs info level formatted message
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warn logs warn level message
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf logs warn level formatted message
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error logs error level message
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf logs error level formatted message
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatal logs fatal level message and exits
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf logs fatal level formatted message and exits
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// SetOutput sets the output destination
func SetOutput(out io.Writer) {
	GetLogger().SetOutput(out)
}

// APILog creates a log entry for API operations
func APILog(handler, method string) *logrus.Entry {
	return WithFields(Fields{
		"handler": handler,
		"method":  method,
	})
}

// ServiceLog creates a log entry for service operations
func ServiceLog(service, method string) *logrus.Entry {
	return WithFields(Fields{
		"service": service,
		"method":  method,
	})
}

// RepoLog creates a log entry for repository operations
func RepoLog(repo, method string) *logrus.Entry {
	return WithFields(Fields{
		"repository": repo,
		"method":     method,
	})
}

// RequestLog creates a log entry from gin context
func RequestLog(c *gin.Context) *logrus.Entry {
	return WithFields(Fields{
		"request_id": c.GetString("request_id"),
		"client_ip":  c.ClientIP(),
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"user_agent": c.Request.UserAgent(),
	})
}

// UserLog creates a log entry with user context
func UserLog(userID uint) *logrus.Entry {
	return WithField("user_id", userID)
}
