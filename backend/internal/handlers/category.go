package handlers

import (
	"billing-note/internal/repository"
	"billing-note/pkg/errors"
	"billing-note/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryHandler(categoryRepo repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{categoryRepo: categoryRepo}
}

func (h *CategoryHandler) GetAll(c *gin.Context) {
	log := logger.APILog("CategoryHandler", "GetAll")
	requestID := c.GetString("request_id")

	log.WithField("request_id", requestID).Debug("Fetching all categories")

	categories, err := h.categoryRepo.GetAll()
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to fetch categories")

		appErr := errors.NewInternalError("Failed to retrieve categories", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"count":      len(categories),
	}).Debug("Categories retrieved successfully")

	c.JSON(http.StatusOK, categories)
}

func (h *CategoryHandler) GetByType(c *gin.Context) {
	log := logger.APILog("CategoryHandler", "GetByType")
	requestID := c.GetString("request_id")

	categoryType := c.Param("type")

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"type":       categoryType,
	}).Debug("Fetching categories by type")

	if categoryType != "income" && categoryType != "expense" {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"type":       categoryType,
		}).Warn("Invalid category type requested")

		appErr := errors.NewInvalidInputError("type", "must be 'income' or 'expense'")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	categories, err := h.categoryRepo.GetByType(categoryType)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"type":       categoryType,
			"error":      err.Error(),
		}).Error("Failed to fetch categories by type")

		appErr := errors.NewInternalError("Failed to retrieve categories", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"type":       categoryType,
		"count":      len(categories),
	}).Debug("Categories by type retrieved successfully")

	c.JSON(http.StatusOK, categories)
}
