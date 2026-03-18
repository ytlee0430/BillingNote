package models

import "time"

type Budget struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	CategoryID    uint      `gorm:"not null" json:"category_id"`
	MonthlyAmount float64   `gorm:"type:decimal(10,2);not null" json:"monthly_amount"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	User     User     `gorm:"foreignKey:UserID" json:"-"`
	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (Budget) TableName() string {
	return "budgets"
}

type CreateBudgetRequest struct {
	CategoryID    uint    `json:"category_id" binding:"required"`
	MonthlyAmount float64 `json:"monthly_amount" binding:"required,gt=0"`
}

type UpdateBudgetRequest struct {
	MonthlyAmount float64 `json:"monthly_amount" binding:"required,gt=0"`
}

type BudgetComparison struct {
	Budget       Budget  `json:"budget"`
	ActualAmount float64 `json:"actual_amount"`
	Remaining    float64 `json:"remaining"`
	Percentage   float64 `json:"percentage"`
	IsOverBudget bool    `json:"is_over_budget"`
}
