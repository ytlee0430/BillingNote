package handlers

import (
	"billing-note/internal/services"
	"billing-note/pkg/errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoryKeywordHandler struct {
	service *services.CategoryKeywordService
}

func NewCategoryKeywordHandler(service *services.CategoryKeywordService) *CategoryKeywordHandler {
	return &CategoryKeywordHandler{service: service}
}

// List returns all keyword rules for the authenticated user
func (h *CategoryKeywordHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")

	keywords, err := h.service.List(userID)
	if err != nil {
		appErr := errors.NewInternalError("Failed to list keyword rules", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(""))
		return
	}

	c.JSON(http.StatusOK, keywords)
}

type addKeywordRequest struct {
	CategoryID uint   `json:"category_id" binding:"required"`
	Keyword    string `json:"keyword" binding:"required"`
}

// Add creates a new keyword rule
func (h *CategoryKeywordHandler) Add(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req addKeywordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errors.NewValidationError(err.Error())
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(""))
		return
	}

	kw, err := h.service.AddKeyword(userID, req.CategoryID, req.Keyword)
	if err != nil {
		appErr := errors.NewInternalError("Failed to add keyword", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(""))
		return
	}

	c.JSON(http.StatusCreated, kw)
}

// Delete removes a keyword rule
func (h *CategoryKeywordHandler) Delete(c *gin.Context) {
	userID := c.GetUint("user_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		appErr := errors.NewValidationError("Invalid keyword ID")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(""))
		return
	}

	if err := h.service.DeleteKeyword(uint(id), userID); err != nil {
		appErr := errors.NewInternalError("Failed to delete keyword", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(""))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Keyword deleted"})
}

type batchSetRequest struct {
	CategoryID uint     `json:"category_id" binding:"required"`
	Keywords   []string `json:"keywords" binding:"required"`
}

// BatchSet replaces all keywords for a given category
func (h *CategoryKeywordHandler) BatchSet(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req batchSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errors.NewValidationError(err.Error())
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(""))
		return
	}

	if err := h.service.BatchSet(userID, req.CategoryID, req.Keywords); err != nil {
		appErr := errors.NewInternalError("Failed to update keywords", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(""))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Keywords updated"})
}

// Reclassify applies keyword rules to all uncategorized transactions
func (h *CategoryKeywordHandler) Reclassify(c *gin.Context) {
	userID := c.GetUint("user_id")

	updated, err := h.service.ReclassifyAll(userID)
	if err != nil {
		appErr := errors.NewInternalError("Failed to reclassify transactions", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(""))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reclassification complete", "updated": updated})
}

// InitDefaults seeds default keyword rules for the user
func (h *CategoryKeywordHandler) InitDefaults(c *gin.Context) {
	userID := c.GetUint("user_id")

	if err := h.service.InitDefaults(userID); err != nil {
		appErr := errors.NewInternalError("Failed to initialize defaults", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(""))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default keywords initialized"})
}
