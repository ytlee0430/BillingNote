package models

import (
	"time"
)

// PDFPassword stores encrypted PDF passwords for users
type PDFPassword struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	UserID            uint      `gorm:"not null;index" json:"user_id"`
	PasswordEncrypted string    `gorm:"not null" json:"-"`
	Priority          int       `gorm:"not null" json:"priority"` // 1-4
	Label             string    `json:"label"`                    // Optional label like "身分證後4碼"
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName specifies the table name
func (PDFPassword) TableName() string {
	return "user_pdf_passwords"
}

// PDFPasswordInput represents input for creating/updating PDF password
type PDFPasswordInput struct {
	Password string `json:"password" binding:"required"`
	Priority int    `json:"priority" binding:"required,min=1,max=4"`
	Label    string `json:"label"`
}

// PDFPasswordResponse represents the response for PDF password (without actual password)
type PDFPasswordResponse struct {
	ID        uint      `json:"id"`
	Priority  int       `json:"priority"`
	Label     string    `json:"label"`
	HasValue  bool      `json:"has_value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts PDFPassword to PDFPasswordResponse
func (p *PDFPassword) ToResponse() PDFPasswordResponse {
	return PDFPasswordResponse{
		ID:        p.ID,
		Priority:  p.Priority,
		Label:     p.Label,
		HasValue:  p.PasswordEncrypted != "",
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
