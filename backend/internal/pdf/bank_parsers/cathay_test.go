package bank_parsers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCathayParser_BankName(t *testing.T) {
	parser := NewCathayParser()
	assert.Equal(t, "國泰世華", parser.BankName())
}

func TestCathayParser_CanParse(t *testing.T) {
	parser := NewCathayParser()

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "contains 國泰世華",
			content: "這是國泰世華銀行的帳單",
			want:    true,
		},
		{
			name:    "contains CATHAY",
			content: "CATHAY UNITED BANK Statement",
			want:    true,
		},
		{
			name:    "contains 國泰銀行",
			content: "國泰銀行信用卡帳單",
			want:    true,
		},
		{
			name:    "other bank",
			content: "台新銀行信用卡帳單",
			want:    false,
		},
		{
			name:    "empty content",
			content: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.CanParse(tt.content)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCathayParser_Parse(t *testing.T) {
	parser := NewCathayParser()

	content := `
國泰世華銀行信用卡帳單

消費明細：
12/25 全聯福利中心 1,234
12/26 7-ELEVEN 85
12/27 星巴克咖啡 350

總計：1,669
`

	transactions, err := parser.Parse(content)
	require.NoError(t, err)
	assert.Len(t, transactions, 3)

	// Check first transaction
	assert.Equal(t, 12, int(transactions[0].Date.Month()))
	assert.Equal(t, 25, transactions[0].Date.Day())
	assert.Equal(t, "全聯福利中心", transactions[0].Description)
	assert.Equal(t, float64(1234), transactions[0].Amount)
	assert.Equal(t, "TWD", transactions[0].Currency)

	// Check second transaction
	assert.Equal(t, "7-ELEVEN", transactions[1].Description)
	assert.Equal(t, float64(85), transactions[1].Amount)

	// Check third transaction
	assert.Equal(t, "星巴克咖啡", transactions[2].Description)
	assert.Equal(t, float64(350), transactions[2].Amount)
}

func TestCathayParser_ParseEmpty(t *testing.T) {
	parser := NewCathayParser()

	transactions, err := parser.Parse("")
	require.NoError(t, err)
	assert.Empty(t, transactions)
}

func TestCathayParser_ParseYearBoundary(t *testing.T) {
	parser := NewCathayParser()

	// Test that December dates in January are assigned to previous year
	content := `
國泰世華銀行信用卡帳單
12/31 跨年消費 1,000
`
	now := time.Now()
	transactions, err := parser.Parse(content)
	require.NoError(t, err)

	if len(transactions) > 0 {
		// If we're in early part of the year and the transaction is December,
		// it should be from last year
		if now.Month() < 6 {
			assert.Equal(t, now.Year()-1, transactions[0].Date.Year())
		}
	}
}
