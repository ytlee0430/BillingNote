package repository

import (
	"billing-note/internal/models"
	"time"

	"gorm.io/gorm"
)

// InvoiceRepository defines the interface for invoice data access
type InvoiceRepository interface {
	Create(invoice *models.Invoice) error
	GetByID(id uint) (*models.Invoice, error)
	GetByInvoiceNumber(userID uint, invoiceNumber string) (*models.Invoice, error)
	List(userID uint, startDate, endDate *time.Time, page, pageSize int) ([]models.Invoice, int64, error)
	Update(invoice *models.Invoice) error
	Delete(id uint) error
	BatchCreate(invoices []*models.Invoice) (int, error)
}

type invoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) Create(invoice *models.Invoice) error {
	return r.db.Create(invoice).Error
}

func (r *invoiceRepository) GetByID(id uint) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.db.Preload("DuplicatedTransaction").First(&invoice, id).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) GetByInvoiceNumber(userID uint, invoiceNumber string) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.db.Where("user_id = ? AND invoice_number = ?", userID, invoiceNumber).First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) List(userID uint, startDate, endDate *time.Time, page, pageSize int) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var total int64

	query := r.db.Model(&models.Invoice{}).Where("user_id = ?", userID)

	if startDate != nil {
		query = query.Where("invoice_date >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("invoice_date <= ?", endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	err := query.Preload("DuplicatedTransaction").Order("invoice_date DESC").Find(&invoices).Error
	return invoices, total, err
}

func (r *invoiceRepository) Update(invoice *models.Invoice) error {
	return r.db.Save(invoice).Error
}

func (r *invoiceRepository) Delete(id uint) error {
	return r.db.Delete(&models.Invoice{}, id).Error
}

func (r *invoiceRepository) BatchCreate(invoices []*models.Invoice) (int, error) {
	created := 0
	for _, inv := range invoices {
		// Skip duplicates (by unique constraint)
		var existing models.Invoice
		err := r.db.Where("user_id = ? AND invoice_number = ?", inv.UserID, inv.InvoiceNumber).First(&existing).Error
		if err == nil {
			continue // Already exists
		}
		if err := r.db.Create(inv).Error; err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}
