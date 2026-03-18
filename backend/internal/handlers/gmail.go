package handlers

import (
	"billing-note/internal/middleware"
	"billing-note/internal/models"
	"billing-note/internal/services"
	"billing-note/pkg/errors"
	"billing-note/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GmailHandler handles Gmail integration endpoints
type GmailHandler struct {
	gmailService    *services.GmailService
	gmailScanService *services.GmailScanService
}

// NewGmailHandler creates a new Gmail handler
func NewGmailHandler(gmailService *services.GmailService, scanService *services.GmailScanService) *GmailHandler {
	return &GmailHandler{
		gmailService:    gmailService,
		gmailScanService: scanService,
	}
}

// GetAuthURL returns the Google OAuth authorization URL
// GET /api/gmail/auth
func (h *GmailHandler) GetAuthURL(c *gin.Context) {
	log := logger.APILog("GmailHandler", "GetAuthURL")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	url, err := h.gmailService.GetAuthURL(userID)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to generate Gmail auth URL")
		appErr := errors.NewInternalError("Failed to generate auth URL", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// HandleCallback exchanges the OAuth code for tokens
// POST /api/gmail/callback
func (h *GmailHandler) HandleCallback(c *gin.Context) {
	log := logger.APILog("GmailHandler", "HandleCallback")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	var req struct {
		Code  string `json:"code" binding:"required"`
		State string `json:"state" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request: code and state are required")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	if err := h.gmailService.HandleCallback(userID, req.Code, req.State); err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to handle Gmail OAuth callback")
		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInternalError("Failed to connect Gmail", err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Gmail connected successfully"})
}

// GetStatus returns the Gmail connection status
// GET /api/gmail/status
func (h *GmailHandler) GetStatus(c *gin.Context) {
	log := logger.APILog("GmailHandler", "GetStatus")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	status, err := h.gmailService.GetStatus(userID)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to get Gmail status")
		appErr := errors.NewInternalError("Failed to get Gmail status", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	c.JSON(http.StatusOK, status)
}

// Disconnect removes the Gmail connection
// DELETE /api/gmail/disconnect
func (h *GmailHandler) Disconnect(c *gin.Context) {
	log := logger.APILog("GmailHandler", "Disconnect")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	if err := h.gmailService.Disconnect(userID); err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to disconnect Gmail")
		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInternalError("Failed to disconnect Gmail", err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Gmail disconnected successfully"})
}

// UpdateSettings updates Gmail scan settings
// PUT /api/gmail/settings
func (h *GmailHandler) UpdateSettings(c *gin.Context) {
	log := logger.APILog("GmailHandler", "UpdateSettings")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	var input models.GmailSettingsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		appErr := errors.NewValidationError("Invalid request body: " + err.Error())
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	if err := h.gmailService.UpdateSettings(userID, input); err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to update Gmail settings")
		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInternalError("Failed to update Gmail settings", err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Gmail settings updated"})
}

// GetSettings returns Gmail scan settings
// GET /api/gmail/settings
func (h *GmailHandler) GetSettings(c *gin.Context) {
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	settings, err := h.gmailService.GetSettings(userID)
	if err != nil {
		appErr := errors.NewInternalError("Failed to get Gmail settings", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	c.JSON(http.StatusOK, settings)
}

// TriggerScan triggers a Gmail scan
// POST /api/gmail/scan
func (h *GmailHandler) TriggerScan(c *gin.Context) {
	log := logger.APILog("GmailHandler", "TriggerScan")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	result, err := h.gmailScanService.TriggerScan(userID)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to trigger Gmail scan")
		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInternalError("Failed to scan Gmail", err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	c.JSON(http.StatusOK, result)
}
