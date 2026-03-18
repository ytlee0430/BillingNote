package services

import (
	"billing-note/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Budget Repository ---

type mockBudgetRepo struct {
	mock.Mock
}

func (m *mockBudgetRepo) Create(budget *models.Budget) error {
	args := m.Called(budget)
	return args.Error(0)
}

func (m *mockBudgetRepo) List(userID uint) ([]models.Budget, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Budget), args.Error(1)
}

func (m *mockBudgetRepo) GetByID(id uint) (*models.Budget, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Budget), args.Error(1)
}

func (m *mockBudgetRepo) Update(budget *models.Budget) error {
	args := m.Called(budget)
	return args.Error(0)
}

func (m *mockBudgetRepo) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// --- Tests ---

func TestBudgetService_Create(t *testing.T) {
	budgetRepo := new(mockBudgetRepo)
	txnRepo := new(mockTransactionRepo)
	svc := NewBudgetService(budgetRepo, txnRepo)

	budgetRepo.On("Create", mock.AnythingOfType("*models.Budget")).Return(nil)

	budget, err := svc.Create(1, models.CreateBudgetRequest{
		CategoryID:    5,
		MonthlyAmount: 3000,
	})

	assert.NoError(t, err)
	assert.Equal(t, uint(1), budget.UserID)
	assert.Equal(t, uint(5), budget.CategoryID)
	assert.Equal(t, 3000.0, budget.MonthlyAmount)
}

func TestBudgetService_List(t *testing.T) {
	budgetRepo := new(mockBudgetRepo)
	txnRepo := new(mockTransactionRepo)
	svc := NewBudgetService(budgetRepo, txnRepo)

	budgets := []models.Budget{
		{ID: 1, UserID: 1, CategoryID: 5, MonthlyAmount: 3000},
		{ID: 2, UserID: 1, CategoryID: 6, MonthlyAmount: 5000},
	}
	budgetRepo.On("List", uint(1)).Return(budgets, nil)

	result, err := svc.List(1)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestBudgetService_Update(t *testing.T) {
	budgetRepo := new(mockBudgetRepo)
	txnRepo := new(mockTransactionRepo)
	svc := NewBudgetService(budgetRepo, txnRepo)

	budget := &models.Budget{ID: 1, UserID: 1, CategoryID: 5, MonthlyAmount: 3000}
	budgetRepo.On("GetByID", uint(1)).Return(budget, nil)
	budgetRepo.On("Update", mock.AnythingOfType("*models.Budget")).Return(nil)

	result, err := svc.Update(1, 1, models.UpdateBudgetRequest{MonthlyAmount: 5000})

	assert.NoError(t, err)
	assert.Equal(t, 5000.0, result.MonthlyAmount)
}

func TestBudgetService_Update_Unauthorized(t *testing.T) {
	budgetRepo := new(mockBudgetRepo)
	txnRepo := new(mockTransactionRepo)
	svc := NewBudgetService(budgetRepo, txnRepo)

	budget := &models.Budget{ID: 1, UserID: 2, CategoryID: 5, MonthlyAmount: 3000}
	budgetRepo.On("GetByID", uint(1)).Return(budget, nil)

	_, err := svc.Update(1, 1, models.UpdateBudgetRequest{MonthlyAmount: 5000})

	assert.Error(t, err)
}

func TestBudgetService_Delete(t *testing.T) {
	budgetRepo := new(mockBudgetRepo)
	txnRepo := new(mockTransactionRepo)
	svc := NewBudgetService(budgetRepo, txnRepo)

	budget := &models.Budget{ID: 1, UserID: 1}
	budgetRepo.On("GetByID", uint(1)).Return(budget, nil)
	budgetRepo.On("Delete", uint(1)).Return(nil)

	err := svc.Delete(1, 1)

	assert.NoError(t, err)
}

func TestBudgetService_Compare(t *testing.T) {
	budgetRepo := new(mockBudgetRepo)
	txnRepo := new(mockTransactionRepo)
	svc := NewBudgetService(budgetRepo, txnRepo)

	budgets := []models.Budget{
		{ID: 1, UserID: 1, CategoryID: 5, MonthlyAmount: 3000},
	}
	budgetRepo.On("List", uint(1)).Return(budgets, nil)

	// Transactions for the category
	transactions := []models.Transaction{
		{ID: 10, Amount: 1000, Type: "expense", TransactionDate: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)},
		{ID: 11, Amount: 500, Type: "expense", TransactionDate: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)},
		{ID: 12, Amount: 200, Type: "income", TransactionDate: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)}, // income, should not count
	}
	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return(transactions, int64(3), nil)

	comparisons, err := svc.Compare(1, 2026, 1)

	assert.NoError(t, err)
	assert.Len(t, comparisons, 1)
	assert.Equal(t, 1500.0, comparisons[0].ActualAmount) // 1000 + 500, income excluded
	assert.Equal(t, 1500.0, comparisons[0].Remaining)
	assert.Equal(t, 50.0, comparisons[0].Percentage)
	assert.False(t, comparisons[0].IsOverBudget)
}

func TestBudgetService_Compare_OverBudget(t *testing.T) {
	budgetRepo := new(mockBudgetRepo)
	txnRepo := new(mockTransactionRepo)
	svc := NewBudgetService(budgetRepo, txnRepo)

	budgets := []models.Budget{
		{ID: 1, UserID: 1, CategoryID: 5, MonthlyAmount: 1000},
	}
	budgetRepo.On("List", uint(1)).Return(budgets, nil)

	transactions := []models.Transaction{
		{ID: 10, Amount: 800, Type: "expense"},
		{ID: 11, Amount: 500, Type: "expense"},
	}
	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return(transactions, int64(2), nil)

	comparisons, err := svc.Compare(1, 2026, 1)

	assert.NoError(t, err)
	assert.True(t, comparisons[0].IsOverBudget)
	assert.Equal(t, 1300.0, comparisons[0].ActualAmount)
	assert.Equal(t, -300.0, comparisons[0].Remaining)
}
