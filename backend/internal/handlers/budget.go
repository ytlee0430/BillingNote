package handlers

import (
	"billing-note/internal/middleware"
	"billing-note/internal/models"
	"billing-note/internal/services"
	"billing-note/pkg/errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type BudgetHandler struct {
	budgetService *services.BudgetService
}

func NewBudgetHandler(budgetService *services.BudgetService) *BudgetHandler {
	return &BudgetHandler{budgetService: budgetService}
}

func (h *BudgetHandler) Create(c *gin.Context) {
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	var req models.CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request: category_id and monthly_amount are required")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	budget, err := h.budgetService.Create(userID, req)
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInternalError("Failed to create budget", err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	c.JSON(http.StatusCreated, budget)
}

func (h *BudgetHandler) List(c *gin.Context) {
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	budgets, err := h.budgetService.List(userID)
	if err != nil {
		appErr := errors.NewInternalError("Failed to list budgets", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	c.JSON(http.StatusOK, gin.H{"budgets": budgets})
}

func (h *BudgetHandler) Update(c *gin.Context) {
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		appErr := errors.NewValidationError("Invalid budget ID")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	var req models.UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request: monthly_amount is required")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	budget, err := h.budgetService.Update(uint(id), userID, req)
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInternalError("Failed to update budget", err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	c.JSON(http.StatusOK, budget)
}

func (h *BudgetHandler) Delete(c *gin.Context) {
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		appErr := errors.NewValidationError("Invalid budget ID")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	if err := h.budgetService.Delete(uint(id), userID); err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInternalError("Failed to delete budget", err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Budget deleted"})
}

func (h *BudgetHandler) Compare(c *gin.Context) {
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	now := time.Now()
	yearStr := c.DefaultQuery("year", strconv.Itoa(now.Year()))
	monthStr := c.DefaultQuery("month", strconv.Itoa(int(now.Month())))
	year, _ := strconv.Atoi(yearStr)
	month, _ := strconv.Atoi(monthStr)

	comparisons, err := h.budgetService.Compare(userID, year, month)
	if err != nil {
		appErr := errors.NewInternalError("Failed to compare budgets", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	c.JSON(http.StatusOK, gin.H{"comparisons": comparisons})
}
