package handlers

import (
	"billing-note/internal/middleware"
	"billing-note/internal/models"
	"billing-note/internal/services"
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
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	passwords, err := h.passwordService.GetPasswords(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"passwords": passwords,
	})
}

// Set sets or updates a PDF password
// POST /api/settings/pdf-passwords
func (h *PDFPasswordHandler) Set(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input models.PDFPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.passwordService.SetPassword(userID, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req SetMultipleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.passwordService.SetMultiplePasswords(userID, req.Passwords); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "passwords saved successfully",
	})
}

// Delete deletes a PDF password
// DELETE /api/settings/pdf-passwords/:priority
func (h *PDFPasswordHandler) Delete(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	priorityStr := c.Param("priority")
	priority, err := strconv.Atoi(priorityStr)
	if err != nil || priority < 1 || priority > 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid priority (must be 1-4)"})
		return
	}

	if err := h.passwordService.DeletePassword(userID, priority); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "password deleted successfully",
	})
}
