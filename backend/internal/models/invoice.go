package models

import (
	"encoding/json"
	"time"
)

// Invoice represents a cloud invoice synced from MOF API
type Invoice struct {
	ID                      uint             `gorm:"primaryKey" json:"id"`
	UserID                  uint             `gorm:"not null;index" json:"user_id"`
	InvoiceNumber           string           `gorm:"not null;size:10" json:"invoice_number"`
	InvoiceDate             time.Time        `gorm:"not null" json:"invoice_date"`
	SellerName              string           `gorm:"size:255" json:"seller_name"`
	SellerBAN               string           `gorm:"size:8;column:seller_ban" json:"seller_ban"`
	Amount                  float64          `gorm:"type:decimal(10,2);not null" json:"amount"`
	Status                  string           `gorm:"size:50" json:"status"`
	Items                   json.RawMessage  `gorm:"type:jsonb" json:"items"`
	IsDuplicated            bool             `gorm:"default:false" json:"is_duplicated"`
	DuplicatedTransactionID *uint            `json:"duplicated_transaction_id,omitempty"`
	ConfidenceScore         *float64         `gorm:"type:decimal(3,2)" json:"confidence_score,omitempty"`
	CreatedAt               time.Time        `json:"created_at"`

	User                    User             `gorm:"foreignKey:UserID" json:"-"`
	DuplicatedTransaction   *Transaction     `gorm:"foreignKey:DuplicatedTransactionID" json:"duplicated_transaction,omitempty"`
}

func (Invoice) TableName() string {
	return "invoices"
}

// InvoiceItem represents an item in an invoice
type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    string  `json:"quantity"`
	UnitPrice   string  `json:"unit_price"`
	Amount      string  `json:"amount"`
}

// InvoiceSyncRequest represents the request to sync invoices
type InvoiceSyncRequest struct {
	StartDate string `json:"start_date" binding:"required"` // YYYY/MM/DD
	EndDate   string `json:"end_date" binding:"required"`   // YYYY/MM/DD
}

// InvoiceSettingsInput represents input for updating invoice settings
type InvoiceSettingsInput struct {
	InvoiceCarrier string `json:"invoice_carrier" binding:"required"`
}

// ConfirmDuplicateRequest represents input for confirming a duplicate
type ConfirmDuplicateRequest struct {
	InvoiceID     uint `json:"invoice_id" binding:"required"`
	TransactionID uint `json:"transaction_id" binding:"required"`
}
