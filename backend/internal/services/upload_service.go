package services

import (
	"billing-note/internal/models"
	"billing-note/internal/pdf"
	"billing-note/internal/pdf/bank_parsers"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

// UploadService handles PDF upload and parsing
type UploadService struct {
	db              *gorm.DB
	passwordService *PDFPasswordService
	uploadDir       string
	registry        *pdf.ParserRegistry
}

// ParsedTransaction represents a transaction parsed from PDF
type ParsedTransaction struct {
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Category    string    `json:"category"`
	CardLast4   string    `json:"card_last4"`
	IsDuplicate bool      `json:"is_duplicate"`
}

// UploadResult represents the result of PDF upload and parsing
type UploadResult struct {
	Filename     string              `json:"filename"`
	Bank         string              `json:"bank"`
	Transactions []ParsedTransaction `json:"transactions"`
	TotalAmount  float64             `json:"total_amount"`
	Error        string              `json:"error,omitempty"`
}

// NewUploadService creates a new upload service
func NewUploadService(db *gorm.DB, passwordService *PDFPasswordService, uploadDir string) *UploadService {
	registry := bank_parsers.NewRegistryWithAllParsers()

	return &UploadService{
		db:              db,
		passwordService: passwordService,
		uploadDir:       uploadDir,
		registry:        registry,
	}
}

// SaveUploadedFile saves an uploaded file and returns the file path
func (s *UploadService) SaveUploadedFile(userID uint, file *multipart.FileHeader) (string, error) {
	// Create directory structure: uploads/{user_id}/pdfs/{year}/{month}/
	now := time.Now()
	dir := filepath.Join(s.uploadDir, fmt.Sprintf("%d", userID), "pdfs",
		fmt.Sprintf("%d", now.Year()), fmt.Sprintf("%02d", now.Month()))

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
	filePath := filepath.Join(dir, filename)

	// Open source file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy content
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filePath, nil
}

// ParsePDF parses a PDF file and returns transactions
func (s *UploadService) ParsePDF(userID uint, filePath string) (*UploadResult, error) {
	filename := filepath.Base(filePath)

	// Get user's passwords
	passwords, err := s.passwordService.GetDecryptedPasswords(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get passwords: %w", err)
	}

	// Also try passwords from file name rules
	rulePasswords := s.registry.GetPasswordsForFile(filename)
	passwords = append(passwords, rulePasswords...)

	// Parse the PDF
	transactions, bankName, err := s.registry.Parse(filePath, passwords)
	if err != nil {
		return &UploadResult{
			Filename: filename,
			Error:    err.Error(),
		}, nil
	}

	// Convert to ParsedTransaction and check for duplicates
	parsedTransactions := make([]ParsedTransaction, len(transactions))
	totalAmount := 0.0

	for i, t := range transactions {
		isDuplicate := s.checkDuplicate(userID, t)
		parsedTransactions[i] = ParsedTransaction{
			Date:        t.Date,
			Description: t.Description,
			Amount:      t.Amount,
			Currency:    t.Currency,
			Category:    t.Category,
			CardLast4:   t.CardLast4,
			IsDuplicate: isDuplicate,
		}
		totalAmount += t.Amount
	}

	return &UploadResult{
		Filename:     filename,
		Bank:         bankName,
		Transactions: parsedTransactions,
		TotalAmount:  totalAmount,
	}, nil
}

// checkDuplicate checks if a transaction already exists
func (s *UploadService) checkDuplicate(userID uint, t pdf.Transaction) bool {
	var count int64
	s.db.Model(&models.Transaction{}).
		Where("user_id = ? AND transaction_date = ? AND amount = ? AND description = ?",
			userID, t.Date, t.Amount, t.Description).
		Count(&count)
	return count > 0
}

// ImportTransactions imports parsed transactions to database
func (s *UploadService) ImportTransactions(userID uint, transactions []ParsedTransaction) (int, error) {
	imported := 0

	for _, t := range transactions {
		if t.IsDuplicate {
			continue
		}

		transaction := models.Transaction{
			UserID:          userID,
			TransactionDate: t.Date,
			Description:     t.Description,
			Amount:          t.Amount,
			Type:            "expense",
			Source:          "pdf_import",
		}

		// Try to find or create category
		if t.Category != "" {
			var category models.Category
			if err := s.db.Where("user_id = ? AND name = ?", userID, t.Category).First(&category).Error; err == nil {
				transaction.CategoryID = &category.ID
			}
		}

		if err := s.db.Create(&transaction).Error; err != nil {
			return imported, fmt.Errorf("failed to import transaction: %w", err)
		}
		imported++
	}

	return imported, nil
}

// ParseMultiplePDFs parses multiple PDF files
func (s *UploadService) ParseMultiplePDFs(userID uint, filePaths []string) ([]UploadResult, error) {
	results := make([]UploadResult, len(filePaths))

	for i, path := range filePaths {
		result, err := s.ParsePDF(userID, path)
		if err != nil {
			results[i] = UploadResult{
				Filename: filepath.Base(path),
				Error:    err.Error(),
			}
		} else {
			results[i] = *result
		}
	}

	return results, nil
}
