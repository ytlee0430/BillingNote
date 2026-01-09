package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCategory_TableName(t *testing.T) {
	category := Category{}
	assert.Equal(t, "categories", category.TableName())
}

func TestCategory_Fields(t *testing.T) {
	category := Category{
		ID:    1,
		Name:  "餐飲",
		Type:  "expense",
		Icon:  "🍔",
		Color: "#FF6B6B",
	}

	assert.Equal(t, uint(1), category.ID)
	assert.Equal(t, "餐飲", category.Name)
	assert.Equal(t, "expense", category.Type)
	assert.Equal(t, "🍔", category.Icon)
	assert.Equal(t, "#FF6B6B", category.Color)
}

func TestCategory_IncomeType(t *testing.T) {
	category := Category{
		Name: "薪資",
		Type: "income",
	}

	assert.Equal(t, "income", category.Type)
}

func TestCategory_ExpenseType(t *testing.T) {
	category := Category{
		Name: "購物",
		Type: "expense",
	}

	assert.Equal(t, "expense", category.Type)
}
