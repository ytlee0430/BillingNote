package handlers

import (
	"billing-note/internal/middleware"
	"billing-note/internal/services"
	"billing-note/pkg/errors"
	"billing-note/pkg/logger"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// UploadHandler handles PDF upload and parsing
type UploadHandler struct {
	uploadService *services.UploadService
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(uploadService *services.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

// UploadAndParse handles PDF upload and parsing
// POST /api/upload/pdf
func (h *UploadHandler) UploadAndParse(c *gin.Context) {
	log := logger.APILog("UploadHandler", "UploadAndParse")
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
	}).Info("Processing PDF upload request")

	// Get uploaded files
	form, err := c.MultipartForm()
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Warn("Failed to parse multipart form")

		appErr := errors.NewValidationError("Failed to parse form data: " + err.Error())
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
		}).Warn("No files uploaded")

		appErr := errors.NewValidationError("No files uploaded. Please select PDF files to upload.")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
		"file_count": len(files),
	}).Info("Processing uploaded files")

	results := make([]services.UploadResult, 0, len(files))

	for _, file := range files {
		fileLog := log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"filename":   file.Filename,
			"size":       file.Size,
		})

		// Check file type by extension and content type
		ext := strings.ToLower(filepath.Ext(file.Filename))
		contentType := file.Header.Get("Content-Type")
		isPDF := ext == ".pdf" ||
			contentType == "application/pdf" ||
			strings.Contains(contentType, "pdf")

		if !isPDF {
			fileLog.WithField("content_type", contentType).Warn("File is not a PDF")
			results = append(results, services.UploadResult{
				Filename: file.Filename,
				Error:    "File is not a PDF. Only PDF files are supported.",
			})
			continue
		}

		fileLog.Debug("Saving uploaded PDF file")

		// Save file
		filePath, err := h.uploadService.SaveUploadedFile(userID, file)
		if err != nil {
			fileLog.WithError(err).Error("Failed to save uploaded file")
			results = append(results, services.UploadResult{
				Filename: file.Filename,
				Error:    "Failed to save file: " + err.Error(),
			})
			continue
		}

		fileLog.WithField("file_path", filePath).Debug("File saved, starting PDF parsing")

		// Parse PDF
		result, err := h.uploadService.ParsePDF(userID, filePath)
		if err != nil {
			fileLog.WithError(err).Error("Failed to parse PDF")
			results = append(results, services.UploadResult{
				Filename: file.Filename,
				Error:    "Failed to parse PDF: " + err.Error(),
			})
			continue
		}

		fileLog.WithFields(logger.Fields{
			"bank":              result.Bank,
			"transaction_count": len(result.Transactions),
			"total_amount":      result.TotalAmount,
		}).Info("PDF parsed successfully")

		results = append(results, *result)
	}

	// Summary logging
	successCount := 0
	errorCount := 0
	for _, r := range results {
		if r.Error == "" {
			successCount++
		} else {
			errorCount++
		}
	}

	log.WithFields(logger.Fields{
		"request_id":    requestID,
		"user_id":       userID,
		"total_files":   len(files),
		"success_count": successCount,
		"error_count":   errorCount,
	}).Info("PDF upload processing completed")

	c.JSON(http.StatusOK, gin.H{
		"results": results,
	})
}

// ImportRequest represents request for importing transactions
type ImportRequest struct {
	Transactions []services.ParsedTransaction `json:"transactions" binding:"required"`
}

// Import handles importing parsed transactions
// POST /api/transactions/import
func (h *UploadHandler) Import(c *gin.Context) {
	log := logger.APILog("UploadHandler", "Import")
	requestID := c.GetString("request_id")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		log.WithField("request_id", requestID).Warn("Unauthorized: user not authenticated")
		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	var req ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Warn("Invalid request body for transaction import")

		appErr := errors.NewValidationError("Invalid request body: " + err.Error())
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	log.WithFields(logger.Fields{
		"request_id":        requestID,
		"user_id":           userID,
		"transaction_count": len(req.Transactions),
	}).Info("Importing transactions from PDF")

	imported, err := h.uploadService.ImportTransactions(userID, req.Transactions)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to import transactions")

		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInternalError("Failed to import transactions", err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	log.WithFields(logger.Fields{
		"request_id":      requestID,
		"user_id":         userID,
		"imported_count":  imported,
		"submitted_count": len(req.Transactions),
		"duplicate_count": len(req.Transactions) - imported,
	}).Info("Transactions imported successfully")

	c.JSON(http.StatusOK, gin.H{
		"imported": imported,
		"message":  "transactions imported successfully",
	})
}
