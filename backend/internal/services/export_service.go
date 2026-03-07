package services

import (
	"billing-note/internal/repository"
	"bytes"
	"encoding/csv"
	"fmt"
	"time"
)

type ExportService struct {
	transactionRepo repository.TransactionRepository
}

func NewExportService(transactionRepo repository.TransactionRepository) *ExportService {
	return &ExportService{transactionRepo: transactionRepo}
}

func (s *ExportService) ExportCSV(userID uint, startDate, endDate time.Time) ([]byte, error) {
	filter := repository.TransactionFilter{
		UserID:    userID,
		StartDate: &startDate,
		EndDate:   &endDate,
		Page:      0,
		PageSize:  0,
	}

	transactions, _, err := s.transactionRepo.List(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Header
	header := []string{"Date", "Type", "Category", "Description", "Amount", "Source"}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Rows
	for _, txn := range transactions {
		categoryName := ""
		if txn.Category != nil {
			categoryName = txn.Category.Name
		}

		row := []string{
			txn.TransactionDate.Format("2006-01-02"),
			txn.Type,
			categoryName,
			txn.Description,
			fmt.Sprintf("%.2f", txn.Amount),
			txn.Source,
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}
