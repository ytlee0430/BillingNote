package handlers

import (
	"billing-note/internal/middleware"
	"billing-note/internal/services"
	"billing-note/pkg/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SharingHandler struct {
	sharingService *services.SharingService
}

func NewSharingHandler(sharingService *services.SharingService) *SharingHandler {
	return &SharingHandler{sharingService: sharingService}
}

func (h *SharingHandler) GetMyCode(c *gin.Context) {
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	code, err := h.sharingService.GetOrCreateCode(userID)
	if err != nil {
		appErr := errors.NewInternalError("Failed to get pairing code", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": code.Code})
}

func (h *SharingHandler) RegenerateCode(c *gin.Context) {
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	code, err := h.sharingService.RegenerateCode(userID)
	if err != nil {
		appErr := errors.NewInternalError("Failed to regenerate pairing code", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": code.Code})
}

type pairRequest struct {
	Code string `json:"code" binding:"required"`
}

func (h *SharingHandler) Pair(c *gin.Context) {
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	var req pairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request: code is required")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	if err := h.sharingService.Pair(userID, req.Code); err != nil {
		appErr := errors.NewValidationError(err.Error())
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "paired successfully"})
}

func (h *SharingHandler) ListViewers(c *gin.Context) {
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	viewers, err := h.sharingService.ListViewers(userID)
	if err != nil {
		appErr := errors.NewInternalError("Failed to list viewers", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	owners, err := h.sharingService.ListOwners(userID)
	if err != nil {
		appErr := errors.NewInternalError("Failed to list owners", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"viewers": viewers,
		"owners":  owners,
	})
}
