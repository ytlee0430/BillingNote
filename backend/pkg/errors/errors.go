package errors

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

// AppError represents an application error with context
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Detail     string `json:"detail,omitempty"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
	Stack      string `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetail adds detail to the error
func (e *AppError) WithDetail(detail string) *AppError {
	e.Detail = detail
	return e
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// getStack captures the current stack trace
func getStack() string {
	buf := make([]byte, 2048)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// Error codes
const (
	// Authentication errors
	ErrCodeUnauthorized     = "AUTH_UNAUTHORIZED"
	ErrCodeInvalidToken     = "AUTH_INVALID_TOKEN"
	ErrCodeTokenExpired     = "AUTH_TOKEN_EXPIRED"
	ErrCodeInvalidCredentials = "AUTH_INVALID_CREDENTIALS"

	// Validation errors
	ErrCodeValidation       = "VALIDATION_ERROR"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeMissingField     = "MISSING_REQUIRED_FIELD"

	// Resource errors
	ErrCodeNotFound         = "RESOURCE_NOT_FOUND"
	ErrCodeConflict         = "RESOURCE_CONFLICT"
	ErrCodeForbidden        = "RESOURCE_FORBIDDEN"

	// Database errors
	ErrCodeDBError          = "DATABASE_ERROR"
	ErrCodeDBConnection     = "DATABASE_CONNECTION_ERROR"
	ErrCodeDBQuery          = "DATABASE_QUERY_ERROR"

	// File errors
	ErrCodeFileUpload       = "FILE_UPLOAD_ERROR"
	ErrCodeFileNotFound     = "FILE_NOT_FOUND"
	ErrCodeFileParse        = "FILE_PARSE_ERROR"
	ErrCodeInvalidFileType  = "INVALID_FILE_TYPE"

	// Service errors
	ErrCodeInternal         = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"

	// Encryption errors
	ErrCodeEncryption       = "ENCRYPTION_ERROR"
	ErrCodeDecryption       = "DECRYPTION_ERROR"
)

// Factory functions for common errors

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeUnauthorized,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
		Stack:      getStack(),
	}
}

// NewInvalidTokenError creates an invalid token error
func NewInvalidTokenError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeInvalidToken,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
		Stack:      getStack(),
	}
}

// NewInvalidCredentialsError creates an invalid credentials error
func NewInvalidCredentialsError() *AppError {
	return &AppError{
		Code:       ErrCodeInvalidCredentials,
		Message:    "Invalid email or password",
		HTTPStatus: http.StatusUnauthorized,
		Stack:      getStack(),
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeValidation,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
		Stack:      getStack(),
	}
}

// NewInvalidInputError creates an invalid input error
func NewInvalidInputError(field, reason string) *AppError {
	return &AppError{
		Code:       ErrCodeInvalidInput,
		Message:    fmt.Sprintf("Invalid input for field '%s': %s", field, reason),
		HTTPStatus: http.StatusBadRequest,
		Stack:      getStack(),
	}
}

// NewMissingFieldError creates a missing field error
func NewMissingFieldError(field string) *AppError {
	return &AppError{
		Code:       ErrCodeMissingField,
		Message:    fmt.Sprintf("Required field '%s' is missing", field),
		HTTPStatus: http.StatusBadRequest,
		Stack:      getStack(),
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string, id interface{}) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    fmt.Sprintf("%s with ID '%v' not found", resource, id),
		HTTPStatus: http.StatusNotFound,
		Stack:      getStack(),
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeConflict,
		Message:    message,
		HTTPStatus: http.StatusConflict,
		Stack:      getStack(),
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeForbidden,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
		Stack:      getStack(),
	}
}

// NewDBError creates a database error
func NewDBError(operation string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeDBError,
		Message:    fmt.Sprintf("Database error during %s", operation),
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
		Stack:      getStack(),
	}
}

// NewDBQueryError creates a database query error
func NewDBQueryError(query string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeDBQuery,
		Message:    fmt.Sprintf("Database query failed: %s", query),
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
		Stack:      getStack(),
	}
}

// NewFileUploadError creates a file upload error
func NewFileUploadError(filename string, reason string) *AppError {
	return &AppError{
		Code:       ErrCodeFileUpload,
		Message:    fmt.Sprintf("Failed to upload file '%s': %s", filename, reason),
		HTTPStatus: http.StatusBadRequest,
		Stack:      getStack(),
	}
}

// NewFileParseError creates a file parse error
func NewFileParseError(filename string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeFileParse,
		Message:    fmt.Sprintf("Failed to parse file '%s'", filename),
		HTTPStatus: http.StatusUnprocessableEntity,
		Err:        err,
		Stack:      getStack(),
	}
}

// NewInvalidFileTypeError creates an invalid file type error
func NewInvalidFileTypeError(filename, expectedType string) *AppError {
	return &AppError{
		Code:       ErrCodeInvalidFileType,
		Message:    fmt.Sprintf("File '%s' is not a valid %s", filename, expectedType),
		HTTPStatus: http.StatusBadRequest,
		Stack:      getStack(),
	}
}

// NewInternalError creates an internal server error
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeInternal,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
		Stack:      getStack(),
	}
}

// NewEncryptionError creates an encryption error
func NewEncryptionError(operation string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeEncryption,
		Message:    fmt.Sprintf("Encryption error during %s", operation),
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
		Stack:      getStack(),
	}
}

// NewDecryptionError creates a decryption error
func NewDecryptionError(err error) *AppError {
	return &AppError{
		Code:       ErrCodeDecryption,
		Message:    "Failed to decrypt data",
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
		Stack:      getStack(),
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError extracts AppError from an error
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}

// WrapError wraps a regular error into an AppError
func WrapError(err error, code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
		Stack:      getStack(),
	}
}

// ErrorResponse is the JSON response format for errors
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Detail  string `json:"detail,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

// ToResponse converts an AppError to an ErrorResponse
func (e *AppError) ToResponse(traceID string) ErrorResponse {
	return ErrorResponse{
		Error:   e.Message,
		Code:    e.Code,
		Detail:  e.Detail,
		TraceID: traceID,
	}
}

// LogFields returns fields for structured logging
func (e *AppError) LogFields() map[string]interface{} {
	fields := map[string]interface{}{
		"error_code":    e.Code,
		"error_message": e.Message,
		"http_status":   e.HTTPStatus,
	}
	if e.Detail != "" {
		fields["error_detail"] = e.Detail
	}
	if e.Err != nil {
		fields["underlying_error"] = e.Err.Error()
	}
	// Include shortened stack trace (first few lines)
	if e.Stack != "" {
		lines := strings.Split(e.Stack, "\n")
		if len(lines) > 10 {
			fields["stack_trace"] = strings.Join(lines[:10], "\n")
		} else {
			fields["stack_trace"] = e.Stack
		}
	}
	return fields
}
