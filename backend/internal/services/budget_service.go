package services

import (
	"billing-note/internal/models"
	"billing-note/internal/repository"
	"billing-note/pkg/errors"
	"math"
	"time"
)

type BudgetService struct {
	budgetRepo      repository.BudgetRepository
	transactionRepo repository.TransactionRepository
}

func NewBudgetService(budgetRepo repository.BudgetRepository, transactionRepo repository.TransactionRepository) *BudgetService {
	return &BudgetService{
		budgetRepo:      budgetRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *BudgetService) Create(userID uint, req models.CreateBudgetRequest) (*models.Budget, error) {
	budget := &models.Budget{
		UserID:        userID,
		CategoryID:    req.CategoryID,
		MonthlyAmount: req.MonthlyAmount,
	}
	if err := s.budgetRepo.Create(budget); err != nil {
		return nil, errors.NewDBError("create budget", err)
	}
	return budget, nil
}

func (s *BudgetService) List(userID uint) ([]models.Budget, error) {
	return s.budgetRepo.List(userID)
}

func (s *BudgetService) Update(id, userID uint, req models.UpdateBudgetRequest) (*models.Budget, error) {
	budget, err := s.budgetRepo.GetByID(id)
	if err != nil {
		return nil, errors.NewNotFoundError("Budget", id)
	}
	if budget.UserID != userID {
		return nil, errors.NewUnauthorizedError("Not authorized to update this budget")
	}
	budget.MonthlyAmount = req.MonthlyAmount
	if err := s.budgetRepo.Update(budget); err != nil {
		return nil, errors.NewDBError("update budget", err)
	}
	return budget, nil
}

func (s *BudgetService) Delete(id, userID uint) error {
	budget, err := s.budgetRepo.GetByID(id)
	if err != nil {
		return errors.NewNotFoundError("Budget", id)
	}
	if budget.UserID != userID {
		return errors.NewUnauthorizedError("Not authorized to delete this budget")
	}
	return s.budgetRepo.Delete(id)
}

func (s *BudgetService) Compare(userID uint, year, month int) ([]models.BudgetComparison, error) {
	budgets, err := s.budgetRepo.List(userID)
	if err != nil {
		return nil, err
	}

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	var comparisons []models.BudgetComparison
	for _, budget := range budgets {
		catID := budget.CategoryID
		filter := repository.TransactionFilter{
			UserID:     userID,
			CategoryID: &catID,
			StartDate:  &startDate,
			EndDate:    &endDate,
			Page:       0,
			PageSize:   0,
		}

		transactions, _, err := s.transactionRepo.List(filter)
		if err != nil {
			continue
		}

		var actual float64
		for _, txn := range transactions {
			if txn.Type == "expense" {
				actual += txn.Amount
			}
		}

		remaining := budget.MonthlyAmount - actual
		percentage := 0.0
		if budget.MonthlyAmount > 0 {
			percentage = math.Round(actual/budget.MonthlyAmount*10000) / 100
		}

		comparisons = append(comparisons, models.BudgetComparison{
			Budget:       budget,
			ActualAmount: actual,
			Remaining:    remaining,
			Percentage:   percentage,
			IsOverBudget: actual > budget.MonthlyAmount,
		})
	}

	return comparisons, nil
}
