package repository

import (
	"billing-note/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BudgetRepository interface {
	Create(budget *models.Budget) error
	List(userID uint) ([]models.Budget, error)
	GetByID(id uint) (*models.Budget, error)
	Update(budget *models.Budget) error
	Delete(id uint) error
}

type budgetRepository struct {
	db *gorm.DB
}

func NewBudgetRepository(db *gorm.DB) BudgetRepository {
	return &budgetRepository{db: db}
}

func (r *budgetRepository) Create(budget *models.Budget) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "category_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"monthly_amount", "updated_at"}),
	}).Create(budget).Error
}

func (r *budgetRepository) List(userID uint) ([]models.Budget, error) {
	var budgets []models.Budget
	err := r.db.Preload("Category").Where("user_id = ?", userID).Find(&budgets).Error
	return budgets, err
}

func (r *budgetRepository) GetByID(id uint) (*models.Budget, error) {
	var budget models.Budget
	err := r.db.Preload("Category").First(&budget, id).Error
	if err != nil {
		return nil, err
	}
	return &budget, nil
}

func (r *budgetRepository) Update(budget *models.Budget) error {
	return r.db.Save(budget).Error
}

func (r *budgetRepository) Delete(id uint) error {
	return r.db.Delete(&models.Budget{}, id).Error
}
