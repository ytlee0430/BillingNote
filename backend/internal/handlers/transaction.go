package handlers

import (
	"billing-note/internal/middleware"
	"billing-note/internal/repository"
	"billing-note/internal/services"
	"billing-note/pkg/errors"
	"billing-note/pkg/logger"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	transactionService services.TransactionService
}

func NewTransactionHandler(transactionService services.TransactionService) *TransactionHandler {
	return &TransactionHandler{transactionService: transactionService}
}

func (h *TransactionHandler) Create(c *gin.Context) {
	log := logger.APILog("TransactionHandler", "Create")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		log.WithField("request_id", requestID).Warn("Unauthorized: user not authenticated")
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	var req services.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Warn("Invalid request body for transaction creation")
		appErr := errors.NewValidationError("Invalid request body: " + err.Error())
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
		"amount":     req.Amount,
		"type":       req.Type,
	}).Info("Creating new transaction")

	transaction, err := h.transactionService.CreateTransaction(userID, &req)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to create transaction")

		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			c.JSON(http.StatusBadRequest, errors.ErrorResponse{Error: err.Error(), TraceID: requestID})
		}
		return
	}

	log.WithFields(logger.Fields{
		"request_id":     requestID,
		"user_id":        userID,
		"transaction_id": transaction.ID,
	}).Info("Transaction created successfully")

	c.JSON(http.StatusCreated, transaction)
}

func (h *TransactionHandler) Get(c *gin.Context) {
	log := logger.APILog("TransactionHandler", "Get")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		log.WithField("request_id", requestID).Warn("Unauthorized: user not authenticated")
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"id_param":   idStr,
		}).Warn("Invalid transaction ID format")
		appErr := errors.NewInvalidInputError("id", "must be a positive integer")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	log.WithFields(logger.Fields{
		"request_id":     requestID,
		"user_id":        userID,
		"transaction_id": id,
	}).Debug("Fetching transaction")

	transaction, err := h.transactionService.GetTransaction(uint(id), userID)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id":     requestID,
			"user_id":        userID,
			"transaction_id": id,
			"error":          err.Error(),
		}).Warn("Failed to get transaction")

		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewNotFoundError("Transaction", id)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	log.WithFields(logger.Fields{
		"request_id":     requestID,
		"user_id":        userID,
		"transaction_id": id,
	}).Debug("Transaction retrieved successfully")

	c.JSON(http.StatusOK, transaction)
}

func (h *TransactionHandler) List(c *gin.Context) {
	log := logger.APILog("TransactionHandler", "List")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		log.WithField("request_id", requestID).Warn("Unauthorized: user not authenticated")
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	// Parse query parameters
	filter := repository.TransactionFilter{
		UserID: userID,
	}

	if typeParam := c.Query("type"); typeParam != "" {
		filter.Type = typeParam
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			filter.EndDate = &endDate
		}
	}

	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
		if err == nil {
			catID := uint(categoryID)
			filter.CategoryID = &catID
		}
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	filter.Page = page
	filter.PageSize = pageSize

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
		"page":       page,
		"page_size":  pageSize,
		"type":       filter.Type,
	}).Debug("Listing transactions with filter")

	transactions, total, err := h.transactionService.ListTransactions(userID, filter)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to list transactions")

		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInternalError("Failed to retrieve transactions", err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
		"total":      total,
		"returned":   len(transactions),
	}).Debug("Transactions listed successfully")

	c.JSON(http.StatusOK, gin.H{
		"data":      transactions,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *TransactionHandler) Update(c *gin.Context) {
	log := logger.APILog("TransactionHandler", "Update")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		log.WithField("request_id", requestID).Warn("Unauthorized: user not authenticated")
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"id_param":   idStr,
		}).Warn("Invalid transaction ID format")
		appErr := errors.NewInvalidInputError("id", "must be a positive integer")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	var req services.UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithFields(logger.Fields{
			"request_id":     requestID,
			"user_id":        userID,
			"transaction_id": id,
			"error":          err.Error(),
		}).Warn("Invalid request body for transaction update")
		appErr := errors.NewValidationError("Invalid request body: " + err.Error())
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	log.WithFields(logger.Fields{
		"request_id":     requestID,
		"user_id":        userID,
		"transaction_id": id,
	}).Info("Updating transaction")

	transaction, err := h.transactionService.UpdateTransaction(uint(id), userID, &req)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id":     requestID,
			"user_id":        userID,
			"transaction_id": id,
			"error":          err.Error(),
		}).Error("Failed to update transaction")

		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			c.JSON(http.StatusBadRequest, errors.ErrorResponse{Error: err.Error(), TraceID: requestID})
		}
		return
	}

	log.WithFields(logger.Fields{
		"request_id":     requestID,
		"user_id":        userID,
		"transaction_id": id,
	}).Info("Transaction updated successfully")

	c.JSON(http.StatusOK, transaction)
}

func (h *TransactionHandler) Delete(c *gin.Context) {
	log := logger.APILog("TransactionHandler", "Delete")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		log.WithField("request_id", requestID).Warn("Unauthorized: user not authenticated")
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"id_param":   idStr,
		}).Warn("Invalid transaction ID format")
		appErr := errors.NewInvalidInputError("id", "must be a positive integer")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	log.WithFields(logger.Fields{
		"request_id":     requestID,
		"user_id":        userID,
		"transaction_id": id,
	}).Info("Deleting transaction")

	if err := h.transactionService.DeleteTransaction(uint(id), userID); err != nil {
		log.WithFields(logger.Fields{
			"request_id":     requestID,
			"user_id":        userID,
			"transaction_id": id,
			"error":          err.Error(),
		}).Error("Failed to delete transaction")

		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			c.JSON(http.StatusBadRequest, errors.ErrorResponse{Error: err.Error(), TraceID: requestID})
		}
		return
	}

	log.WithFields(logger.Fields{
		"request_id":     requestID,
		"user_id":        userID,
		"transaction_id": id,
	}).Info("Transaction deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "transaction deleted successfully"})
}

func (h *TransactionHandler) GetMonthlyStats(c *gin.Context) {
	log := logger.APILog("TransactionHandler", "GetMonthlyStats")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		log.WithField("request_id", requestID).Warn("Unauthorized: user not authenticated")
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	monthStr := c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month())))

	year, _ := strconv.Atoi(yearStr)
	month, _ := strconv.Atoi(monthStr)

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
		"year":       year,
		"month":      month,
	}).Debug("Fetching monthly stats")

	stats, err := h.transactionService.GetMonthlyStats(userID, year, month)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"year":       year,
			"month":      month,
			"error":      err.Error(),
		}).Error("Failed to get monthly stats")

		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInternalError("Failed to retrieve monthly stats", err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
		"year":       year,
		"month":      month,
	}).Debug("Monthly stats retrieved successfully")

	c.JSON(http.StatusOK, stats)
}

func (h *TransactionHandler) GetCategoryStats(c *gin.Context) {
	log := logger.APILog("TransactionHandler", "GetCategoryStats")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		log.WithField("request_id", requestID).Warn("Unauthorized: user not authenticated")
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	transactionType := c.Query("type")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"start_date": startDateStr,
		}).Warn("Invalid start_date format")
		appErr := errors.NewInvalidInputError("start_date", "must be in YYYY-MM-DD format")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"end_date":   endDateStr,
		}).Warn("Invalid end_date format")
		appErr := errors.NewInvalidInputError("end_date", "must be in YYYY-MM-DD format")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
		"start_date": startDateStr,
		"end_date":   endDateStr,
		"type":       transactionType,
	}).Debug("Fetching category stats")

	stats, err := h.transactionService.GetCategoryStats(userID, startDate, endDate, transactionType)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to get category stats")

		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInternalError("Failed to retrieve category stats", err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
	}).Debug("Category stats retrieved successfully")

	c.JSON(http.StatusOK, stats)
}
