package handlers

import (
	"billing-note/internal/middleware"
	"billing-note/internal/services"
	"net/http"

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
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get uploaded files
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files uploaded"})
		return
	}

	results := make([]services.UploadResult, 0, len(files))

	for _, file := range files {
		// Check file type
		if file.Header.Get("Content-Type") != "application/pdf" {
			results = append(results, services.UploadResult{
				Filename: file.Filename,
				Error:    "not a PDF file",
			})
			continue
		}

		// Save file
		filePath, err := h.uploadService.SaveUploadedFile(userID, file)
		if err != nil {
			results = append(results, services.UploadResult{
				Filename: file.Filename,
				Error:    err.Error(),
			})
			continue
		}

		// Parse PDF
		result, err := h.uploadService.ParsePDF(userID, filePath)
		if err != nil {
			results = append(results, services.UploadResult{
				Filename: file.Filename,
				Error:    err.Error(),
			})
			continue
		}

		results = append(results, *result)
	}

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
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imported, err := h.uploadService.ImportTransactions(userID, req.Transactions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"imported": imported,
		"message":  "transactions imported successfully",
	})
}
