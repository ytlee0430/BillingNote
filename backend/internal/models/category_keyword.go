package models

import "time"

// CategoryKeyword maps a keyword pattern to a category for auto-classification
type CategoryKeyword struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;index:idx_cat_kw_user" json:"user_id"`
	CategoryID uint      `gorm:"not null;index" json:"category_id"`
	Keyword    string    `gorm:"not null;size:100" json:"keyword"`
	CreatedAt  time.Time `json:"created_at"`

	User     User     `gorm:"foreignKey:UserID" json:"-"`
	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (CategoryKeyword) TableName() string {
	return "category_keywords"
}
