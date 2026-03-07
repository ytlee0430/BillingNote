package handlers

import (
	"billing-note/internal/middleware"
	"billing-note/internal/models"
	"billing-note/internal/services"
	"billing-note/pkg/errors"
	"billing-note/pkg/logger"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InvoiceHandler handles invoice endpoints
type InvoiceHandler struct {
	invoiceService *services.InvoiceService
	db             *gorm.DB
}

// NewInvoiceHandler creates a new invoice handler
func NewInvoiceHandler(invoiceService *services.InvoiceService, db *gorm.DB) *InvoiceHandler {
	return &InvoiceHandler{invoiceService: invoiceService, db: db}
}

// Sync triggers invoice sync from MOF API
// POST /api/invoice/sync
func (h *InvoiceHandler) Sync(c *gin.Context) {
	log := logger.APILog("InvoiceHandler", "Sync")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	var req models.InvoiceSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request: start_date and end_date are required (YYYY/MM/DD)")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	// Get user's carrier code
	var user struct {
		InvoiceCarrier *string
	}
	if err := h.db.Table("users").Select("invoice_carrier").Where("id = ?", userID).Scan(&user).Error; err != nil {
		appErr := errors.NewInternalError("Failed to get user settings", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	if user.InvoiceCarrier == nil || *user.InvoiceCarrier == "" {
		appErr := errors.NewValidationError("Invoice carrier code not set. Please configure it in settings first.")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	created, err := h.invoiceService.SyncInvoices(userID, *user.InvoiceCarrier, req.StartDate, req.EndDate)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to sync invoices")
		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInternalError("Failed to sync invoices", err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"synced":  created,
		"message": "Invoice sync completed",
	})
}

// List returns invoices with pagination
// GET /api/invoice/list
func (h *InvoiceHandler) List(c *gin.Context) {
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var startDate, endDate *time.Time
	if sd := c.Query("start_date"); sd != "" {
		if t, err := time.Parse("2006-01-02", sd); err == nil {
			startDate = &t
		}
	}
	if ed := c.Query("end_date"); ed != "" {
		if t, err := time.Parse("2006-01-02", ed); err == nil {
			endDate = &t
		}
	}

	invoices, total, err := h.invoiceService.ListInvoices(userID, startDate, endDate, page, pageSize)
	if err != nil {
		appErr := errors.NewInternalError("Failed to list invoices", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invoices":  invoices,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ConfirmDuplicate marks an invoice as a confirmed duplicate
// POST /api/invoice/confirm-duplicate
func (h *InvoiceHandler) ConfirmDuplicate(c *gin.Context) {
	requestID := c.GetString("request_id")

	_, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	var req models.ConfirmDuplicateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request: invoice_id and transaction_id are required")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	if err := h.invoiceService.ConfirmDuplicate(req.InvoiceID, req.TransactionID); err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInternalError("Failed to confirm duplicate", err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Duplicate confirmed"})
}

// Delete deletes an invoice
// DELETE /api/invoice/:id
func (h *InvoiceHandler) Delete(c *gin.Context) {
	requestID := c.GetString("request_id")

	_, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		appErr := errors.NewValidationError("Invalid invoice ID")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	if err := h.invoiceService.DeleteInvoice(uint(id)); err != nil {
		appErr := errors.NewInternalError("Failed to delete invoice", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invoice deleted"})
}

// UpdateSettings updates invoice settings (carrier code)
// PUT /api/invoice/settings
func (h *InvoiceHandler) UpdateSettings(c *gin.Context) {
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	var input models.InvoiceSettingsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		appErr := errors.NewValidationError("Invalid request: invoice_carrier is required")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	if err := h.db.Table("users").Where("id = ?", userID).Update("invoice_carrier", input.InvoiceCarrier).Error; err != nil {
		appErr := errors.NewDBError("update invoice settings", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invoice settings updated"})
}
