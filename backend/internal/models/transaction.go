package models

import (
	"time"
)

type Transaction struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"not null;index" json:"user_id"`
	CategoryID      *uint     `gorm:"index" json:"category_id"`
	Amount          float64   `gorm:"not null" json:"amount"`
	Type            string    `gorm:"not null;index" json:"type"` // "income" or "expense"
	Description     string    `json:"description"`
	TransactionDate time.Time `gorm:"not null;index" json:"transaction_date"`
	Source          string    `gorm:"default:manual" json:"source"` // "manual", "pdf", "gmail", "invoice"
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Associations
	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// TableName specifies the table name
func (Transaction) TableName() string {
	return "transactions"
}
