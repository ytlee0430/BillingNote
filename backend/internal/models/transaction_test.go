package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransaction_TableName(t *testing.T) {
	transaction := Transaction{}
	assert.Equal(t, "transactions", transaction.TableName())
}

func TestTransaction_Fields(t *testing.T) {
	now := time.Now()
	categoryID := uint(1)

	transaction := Transaction{
		ID:              1,
		UserID:          100,
		CategoryID:      &categoryID,
		Amount:          150.50,
		Type:            "expense",
		Description:     "Test transaction",
		TransactionDate: now,
		Source:          "manual",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	assert.Equal(t, uint(1), transaction.ID)
	assert.Equal(t, uint(100), transaction.UserID)
	assert.Equal(t, &categoryID, transaction.CategoryID)
	assert.Equal(t, 150.50, transaction.Amount)
	assert.Equal(t, "expense", transaction.Type)
	assert.Equal(t, "Test transaction", transaction.Description)
	assert.Equal(t, "manual", transaction.Source)
	assert.Equal(t, now, transaction.TransactionDate)
}

func TestTransaction_NilCategory(t *testing.T) {
	transaction := Transaction{
		ID:         1,
		UserID:     100,
		CategoryID: nil,
		Amount:     100.0,
		Type:       "income",
	}

	assert.Nil(t, transaction.CategoryID)
}
