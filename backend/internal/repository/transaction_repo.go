package repository

import (
	"billing-note/internal/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

type TransactionFilter struct {
	UserID    uint
	Type      string
	StartDate *time.Time
	EndDate   *time.Time
	CategoryID *uint
	Page      int
	PageSize  int
}

type TransactionRepository interface {
	Create(transaction *models.Transaction) error
	GetByID(id uint) (*models.Transaction, error)
	List(filter TransactionFilter) ([]models.Transaction, int64, error)
	Update(transaction *models.Transaction) error
	Delete(id uint) error
	GetMonthlyStats(userID uint, year int, month int) (map[string]float64, error)
	GetCategoryStats(userID uint, startDate, endDate time.Time, transactionType string) ([]map[string]interface{}, error)
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(transaction *models.Transaction) error {
	return r.db.Create(transaction).Error
}

func (r *transactionRepository) GetByID(id uint) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.db.Preload("Category").First(&transaction, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) List(filter TransactionFilter) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	query := r.db.Model(&models.Transaction{}).Where("user_id = ?", filter.UserID)

	// Apply filters
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.StartDate != nil {
		query = query.Where("transaction_date >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("transaction_date <= ?", filter.EndDate)
	}
	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", *filter.CategoryID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	// Execute query with preloading
	err := query.Preload("Category").Order("transaction_date DESC, id DESC").Find(&transactions).Error
	return transactions, total, err
}

func (r *transactionRepository) Update(transaction *models.Transaction) error {
	return r.db.Save(transaction).Error
}

func (r *transactionRepository) Delete(id uint) error {
	return r.db.Delete(&models.Transaction{}, id).Error
}

func (r *transactionRepository) GetMonthlyStats(userID uint, year int, month int) (map[string]float64, error) {
	stats := make(map[string]float64)

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	// Get income
	var income float64
	if err := r.db.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND transaction_date BETWEEN ? AND ?", userID, "income", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&income).Error; err != nil {
		return nil, err
	}

	// Get expense
	var expense float64
	if err := r.db.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND transaction_date BETWEEN ? AND ?", userID, "expense", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&expense).Error; err != nil {
		return nil, err
	}

	stats["income"] = income
	stats["expense"] = expense
	stats["balance"] = income - expense

	return stats, nil
}

func (r *transactionRepository) GetCategoryStats(userID uint, startDate, endDate time.Time, transactionType string) ([]map[string]interface{}, error) {
	var results []struct {
		CategoryID   *uint
		CategoryName string
		Amount       float64
	}

	query := r.db.Model(&models.Transaction{}).
		Select("transactions.category_id, categories.name as category_name, SUM(transactions.amount) as amount").
		Joins("LEFT JOIN categories ON transactions.category_id = categories.id").
		Where("transactions.user_id = ? AND transactions.transaction_date BETWEEN ? AND ?", userID, startDate, endDate)

	if transactionType != "" {
		query = query.Where("transactions.type = ?", transactionType)
	}

	if err := query.Group("transactions.category_id, categories.name").Scan(&results).Error; err != nil {
		return nil, err
	}

	stats := make([]map[string]interface{}, len(results))
	for i, result := range results {
		categoryName := result.CategoryName
		if categoryName == "" {
			categoryName = "未分類"
		}
		stats[i] = map[string]interface{}{
			"category_id":   result.CategoryID,
			"category_name": categoryName,
			"amount":        result.Amount,
		}
	}

	return stats, nil
}
