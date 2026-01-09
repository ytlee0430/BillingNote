package services

import (
	"billing-note/internal/models"
	"billing-note/internal/repository"
	"errors"
	"time"
)

type TransactionService interface {
	CreateTransaction(userID uint, req *CreateTransactionRequest) (*models.Transaction, error)
	GetTransaction(id uint, userID uint) (*models.Transaction, error)
	ListTransactions(userID uint, filter repository.TransactionFilter) ([]models.Transaction, int64, error)
	UpdateTransaction(id uint, userID uint, req *UpdateTransactionRequest) (*models.Transaction, error)
	DeleteTransaction(id uint, userID uint) error
	GetMonthlyStats(userID uint, year int, month int) (map[string]float64, error)
	GetCategoryStats(userID uint, startDate, endDate time.Time, transactionType string) ([]map[string]interface{}, error)
}

type CreateTransactionRequest struct {
	CategoryID      *uint     `json:"category_id"`
	Amount          float64   `json:"amount" binding:"required,gt=0"`
	Type            string    `json:"type" binding:"required,oneof=income expense"`
	Description     string    `json:"description"`
	TransactionDate time.Time `json:"transaction_date" binding:"required"`
	Source          string    `json:"source"`
}

type UpdateTransactionRequest struct {
	CategoryID      *uint     `json:"category_id"`
	Amount          float64   `json:"amount" binding:"gt=0"`
	Type            string    `json:"type" binding:"oneof=income expense"`
	Description     string    `json:"description"`
	TransactionDate time.Time `json:"transaction_date"`
}

type transactionService struct {
	repo repository.TransactionRepository
}

func NewTransactionService(repo repository.TransactionRepository) TransactionService {
	return &transactionService{repo: repo}
}

func (s *transactionService) CreateTransaction(userID uint, req *CreateTransactionRequest) (*models.Transaction, error) {
	if req.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	if req.Type != "income" && req.Type != "expense" {
		return nil, errors.New("type must be either 'income' or 'expense'")
	}

	source := req.Source
	if source == "" {
		source = "manual"
	}

	transaction := &models.Transaction{
		UserID:          userID,
		CategoryID:      req.CategoryID,
		Amount:          req.Amount,
		Type:            req.Type,
		Description:     req.Description,
		TransactionDate: req.TransactionDate,
		Source:          source,
	}

	if err := s.repo.Create(transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

func (s *transactionService) GetTransaction(id uint, userID uint) (*models.Transaction, error) {
	transaction, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if transaction.UserID != userID {
		return nil, errors.New("unauthorized access to transaction")
	}

	return transaction, nil
}

func (s *transactionService) ListTransactions(userID uint, filter repository.TransactionFilter) ([]models.Transaction, int64, error) {
	filter.UserID = userID
	return s.repo.List(filter)
}

func (s *transactionService) UpdateTransaction(id uint, userID uint, req *UpdateTransactionRequest) (*models.Transaction, error) {
	transaction, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if transaction.UserID != userID {
		return nil, errors.New("unauthorized access to transaction")
	}

	// Update fields if provided
	if req.Amount > 0 {
		transaction.Amount = req.Amount
	}
	if req.Type != "" {
		if req.Type != "income" && req.Type != "expense" {
			return nil, errors.New("type must be either 'income' or 'expense'")
		}
		transaction.Type = req.Type
	}
	if req.CategoryID != nil {
		transaction.CategoryID = req.CategoryID
	}
	if req.Description != "" {
		transaction.Description = req.Description
	}
	if !req.TransactionDate.IsZero() {
		transaction.TransactionDate = req.TransactionDate
	}

	if err := s.repo.Update(transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

func (s *transactionService) DeleteTransaction(id uint, userID uint) error {
	transaction, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if transaction.UserID != userID {
		return errors.New("unauthorized access to transaction")
	}

	return s.repo.Delete(id)
}

func (s *transactionService) GetMonthlyStats(userID uint, year int, month int) (map[string]float64, error) {
	return s.repo.GetMonthlyStats(userID, year, month)
}

func (s *transactionService) GetCategoryStats(userID uint, startDate, endDate time.Time, transactionType string) ([]map[string]interface{}, error) {
	return s.repo.GetCategoryStats(userID, startDate, endDate, transactionType)
}
