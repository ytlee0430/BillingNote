package handlers

import (
	"billing-note/internal/middleware"
	"billing-note/internal/services"
	"billing-note/pkg/errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ExportHandler struct {
	exportService *services.ExportService
}

func NewExportHandler(exportService *services.ExportService) *ExportHandler {
	return &ExportHandler{exportService: exportService}
}

func (h *ExportHandler) ExportCSV(c *gin.Context) {
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		appErr := errors.NewValidationError("Invalid date range: " + err.Error())
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	data, err := h.exportService.ExportCSV(userID, startDate, endDate)
	if err != nil {
		appErr := errors.NewInternalError("Failed to export CSV", err)
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	filename := fmt.Sprintf("transactions_%s_%s.csv",
		startDate.Format("20060102"),
		endDate.Format("20060102"),
	)

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "text/csv", data)
}

func parseDateRange(c *gin.Context) (time.Time, time.Time, error) {
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	if startStr == "" || endStr == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("start_date and end_date are required")
	}

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date format (YYYY-MM-DD)")
	}

	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date format (YYYY-MM-DD)")
	}

	return startDate, endDate, nil
}
