package services

import (
	"billing-note/internal/models"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExportCSV_Success(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	svc := NewExportService(txnRepo)

	category := &models.Category{ID: 1, Name: "Food", Type: "expense"}
	transactions := []models.Transaction{
		{
			ID:              1,
			Amount:          100.50,
			Type:            "expense",
			Description:     "Lunch",
			TransactionDate: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			Source:          "manual",
			Category:        category,
		},
		{
			ID:              2,
			Amount:          5000,
			Type:            "income",
			Description:     "Salary",
			TransactionDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Source:          "manual",
			Category:        nil,
		},
	}

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return(transactions, int64(2), nil)

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	data, err := svc.ExportCSV(1, startDate, endDate)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Parse CSV
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	assert.NoError(t, err)

	// Header + 2 rows
	assert.Len(t, records, 3)
	assert.Equal(t, []string{"Date", "Type", "Category", "Description", "Amount", "Source"}, records[0])
	assert.Equal(t, "2026-01-15", records[1][0])
	assert.Equal(t, "expense", records[1][1])
	assert.Equal(t, "Food", records[1][2])
	assert.Equal(t, "Lunch", records[1][3])
	assert.Equal(t, "100.50", records[1][4])

	// No category
	assert.Equal(t, "", records[2][2])
	assert.Equal(t, "Salary", records[2][3])
}

func TestExportCSV_EmptyTransactions(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	svc := NewExportService(txnRepo)

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return([]models.Transaction{}, int64(0), nil)

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	data, err := svc.ExportCSV(1, startDate, endDate)

	assert.NoError(t, err)

	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	assert.NoError(t, err)
	assert.Len(t, records, 1) // header only
}

func TestExportCSV_RepoError(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	svc := NewExportService(txnRepo)

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return([]models.Transaction{}, int64(0), assert.AnError)

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	_, err := svc.ExportCSV(1, startDate, endDate)

	assert.Error(t, err)
}
