package handlers

import (
	"billing-note/internal/middleware"
	"billing-note/internal/models"
	"billing-note/internal/services"
	"billing-note/pkg/errors"
	"billing-note/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PDFPasswordHandler handles PDF password management
type PDFPasswordHandler struct {
	passwordService *services.PDFPasswordService
}

// NewPDFPasswordHandler creates a new PDF password handler
func NewPDFPasswordHandler(passwordService *services.PDFPasswordService) *PDFPasswordHandler {
	return &PDFPasswordHandler{passwordService: passwordService}
}

// List returns all PDF passwords for the user (without actual passwords)
// GET /api/settings/pdf-passwords
func (h *PDFPasswordHandler) List(c *gin.Context) {
	log := logger.APILog("PDFPasswordHandler", "List")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		log.WithField("request_id", requestID).Warn("Unauthorized: user not authenticated")
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
	}).Debug("Fetching PDF passwords for user")

	passwords, err := h.passwordService.GetPasswords(userID)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to fetch PDF passwords")

		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInternalError("Failed to retrieve PDF passwords", err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
		"count":      len(passwords),
	}).Debug("PDF passwords retrieved successfully")

	c.JSON(http.StatusOK, gin.H{
		"passwords": passwords,
	})
}

// Set sets or updates a PDF password
// POST /api/settings/pdf-passwords
func (h *PDFPasswordHandler) Set(c *gin.Context) {
	log := logger.APILog("PDFPasswordHandler", "Set")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		log.WithField("request_id", requestID).Warn("Unauthorized: user not authenticated")
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	var input models.PDFPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Warn("Invalid request body for PDF password")

		appErr := errors.NewValidationError("Invalid request body: " + err.Error())
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
		"priority":   input.Priority,
		"label":      input.Label,
	}).Info("Setting PDF password")

	if err := h.passwordService.SetPassword(userID, input); err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"priority":   input.Priority,
			"error":      err.Error(),
		}).Error("Failed to set PDF password")

		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInternalError("Failed to save PDF password", err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
		"priority":   input.Priority,
	}).Info("PDF password saved successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "password saved successfully",
	})
}

// SetMultiple sets multiple PDF passwords at once
// PUT /api/settings/pdf-passwords
type SetMultipleRequest struct {
	Passwords []models.PDFPasswordInput `json:"passwords" binding:"required"`
}

func (h *PDFPasswordHandler) SetMultiple(c *gin.Context) {
	log := logger.APILog("PDFPasswordHandler", "SetMultiple")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		log.WithField("request_id", requestID).Warn("Unauthorized: user not authenticated")
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	var req SetMultipleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Warn("Invalid request body for multiple PDF passwords")

		appErr := errors.NewValidationError("Invalid request body: " + err.Error())
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
		"count":      len(req.Passwords),
	}).Info("Setting multiple PDF passwords")

	if err := h.passwordService.SetMultiplePasswords(userID, req.Passwords); err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to set multiple PDF passwords")

		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInternalError("Failed to save PDF passwords", err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
		"count":      len(req.Passwords),
	}).Info("Multiple PDF passwords saved successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "passwords saved successfully",
	})
}

// Delete deletes a PDF password
// DELETE /api/settings/pdf-passwords/:priority
func (h *PDFPasswordHandler) Delete(c *gin.Context) {
	log := logger.APILog("PDFPasswordHandler", "Delete")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		log.WithField("request_id", requestID).Warn("Unauthorized: user not authenticated")
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	priorityStr := c.Param("priority")
	priority, err := strconv.Atoi(priorityStr)
	if err != nil || priority < 1 || priority > 4 {
		log.WithFields(logger.Fields{
			"request_id":     requestID,
			"user_id":        userID,
			"priority_param": priorityStr,
		}).Warn("Invalid priority parameter for PDF password deletion")

		appErr := errors.NewInvalidInputError("priority", "must be a number between 1 and 4")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
		"priority":   priority,
	}).Info("Deleting PDF password")

	if err := h.passwordService.DeletePassword(userID, priority); err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"priority":   priority,
			"error":      err.Error(),
		}).Error("Failed to delete PDF password")

		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewNotFoundError("PDF password", priority)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
		"priority":   priority,
	}).Info("PDF password deleted successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "password deleted successfully",
	})
}
