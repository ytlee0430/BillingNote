package services

import (
	"billing-note/internal/models"
	"billing-note/internal/repository"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListTransactions_WithSearchQuery(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	svc := NewTransactionService(txnRepo)

	transactions := []models.Transaction{
		{
			ID:              1,
			UserID:          1,
			Amount:          50,
			Type:            "expense",
			Description:     "Lunch at McDonalds",
			TransactionDate: time.Now(),
		},
	}

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return(transactions, int64(1), nil).
		Run(func(args mock.Arguments) {
			f := args.Get(0).(repository.TransactionFilter)
			assert.Equal(t, "McDonalds", f.Query)
		})

	filter := repository.TransactionFilter{
		Query: "McDonalds",
	}
	result, total, err := svc.ListTransactions(1, filter)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	txnRepo.AssertExpectations(t)
}

func TestListTransactions_WithTags(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	svc := NewTransactionService(txnRepo)

	transactions := []models.Transaction{
		{
			ID:              1,
			UserID:          1,
			Amount:          100,
			Type:            "expense",
			Description:     "Groceries",
			TransactionDate: time.Now(),
			Tags:            pq.StringArray{"food", "weekly"},
		},
	}

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return(transactions, int64(1), nil).
		Run(func(args mock.Arguments) {
			f := args.Get(0).(repository.TransactionFilter)
			assert.Equal(t, []string{"food"}, f.Tags)
		})

	filter := repository.TransactionFilter{
		Tags: []string{"food"},
	}
	result, total, err := svc.ListTransactions(1, filter)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.Contains(t, []string(result[0].Tags), "food")
	txnRepo.AssertExpectations(t)
}

func TestListTransactions_WithAmountRange(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	svc := NewTransactionService(txnRepo)

	transactions := []models.Transaction{
		{
			ID:     1,
			UserID: 1,
			Amount: 150,
			Type:   "expense",
		},
	}

	min := 100.0
	max := 200.0

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return(transactions, int64(1), nil).
		Run(func(args mock.Arguments) {
			f := args.Get(0).(repository.TransactionFilter)
			assert.Equal(t, &min, f.MinAmount)
			assert.Equal(t, &max, f.MaxAmount)
		})

	filter := repository.TransactionFilter{
		MinAmount: &min,
		MaxAmount: &max,
	}
	result, total, err := svc.ListTransactions(1, filter)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	txnRepo.AssertExpectations(t)
}

func TestCreateTransaction_WithTags(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	svc := NewTransactionService(txnRepo)

	txnRepo.On("Create", mock.AnythingOfType("*models.Transaction")).
		Return(nil).
		Run(func(args mock.Arguments) {
			txn := args.Get(0).(*models.Transaction)
			assert.Equal(t, pq.StringArray{"food", "lunch"}, txn.Tags)
		})

	req := &CreateTransactionRequest{
		Amount:          50,
		Type:            "expense",
		Description:     "Lunch",
		TransactionDate: time.Now(),
		Tags:            []string{"food", "lunch"},
	}

	result, err := svc.CreateTransaction(1, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	txnRepo.AssertExpectations(t)
}

func TestUpdateTransaction_WithTags(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	svc := NewTransactionService(txnRepo)

	existing := &models.Transaction{
		ID:              1,
		UserID:          1,
		Amount:          50,
		Type:            "expense",
		Description:     "Lunch",
		TransactionDate: time.Now(),
		Tags:            pq.StringArray{"food"},
	}

	txnRepo.On("GetByID", uint(1)).Return(existing, nil)
	txnRepo.On("Update", mock.AnythingOfType("*models.Transaction")).
		Return(nil).
		Run(func(args mock.Arguments) {
			txn := args.Get(0).(*models.Transaction)
			assert.Equal(t, pq.StringArray{"food", "lunch"}, pq.StringArray(txn.Tags))
		})

	req := &UpdateTransactionRequest{
		Tags: []string{"food", "lunch"},
	}

	result, err := svc.UpdateTransaction(1, 1, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	txnRepo.AssertExpectations(t)
}

func TestListTransactions_MultiConditionFilter(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	svc := NewTransactionService(txnRepo)

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	min := 50.0
	catID := uint(1)

	transactions := []models.Transaction{
		{
			ID:              1,
			UserID:          1,
			Amount:          100,
			Type:            "expense",
			Description:     "Groceries at Costco",
			TransactionDate: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			Tags:            pq.StringArray{"food"},
		},
	}

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return(transactions, int64(1), nil).
		Run(func(args mock.Arguments) {
			f := args.Get(0).(repository.TransactionFilter)
			assert.Equal(t, "expense", f.Type)
			assert.Equal(t, "Costco", f.Query)
			assert.Equal(t, []string{"food"}, f.Tags)
			assert.Equal(t, &min, f.MinAmount)
			assert.Equal(t, &catID, f.CategoryID)
			assert.NotNil(t, f.StartDate)
			assert.NotNil(t, f.EndDate)
		})

	filter := repository.TransactionFilter{
		Type:       "expense",
		StartDate:  &startDate,
		EndDate:    &endDate,
		CategoryID: &catID,
		Query:      "Costco",
		Tags:       []string{"food"},
		MinAmount:  &min,
	}
	result, total, err := svc.ListTransactions(1, filter)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	txnRepo.AssertExpectations(t)
}
