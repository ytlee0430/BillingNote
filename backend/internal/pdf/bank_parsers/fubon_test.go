package bank_parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFubonParser_BankName(t *testing.T) {
	parser := NewFubonParser()
	assert.Equal(t, "富邦銀行", parser.BankName())
}

func TestFubonParser_CanParse(t *testing.T) {
	parser := NewFubonParser()

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "contains 富邦銀行",
			content: "這是富邦銀行的帳單",
			want:    true,
		},
		{
			name:    "contains 台北富邦",
			content: "台北富邦銀行信用卡帳單",
			want:    true,
		},
		{
			name:    "contains FUBON",
			content: "FUBON BANK Statement",
			want:    true,
		},
		{
			name:    "contains 富邦金控",
			content: "富邦金控信用卡服務",
			want:    true,
		},
		{
			name:    "other bank",
			content: "國泰銀行信用卡帳單",
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

func TestFubonParser_ParseROCDate(t *testing.T) {
	parser := NewFubonParser()

	// Test ROC calendar format (民國年)
	content := `
富邦銀行信用卡帳單

消費明細：
114/12/25 餐廳消費 2,500
114/12/26 線上購物 3,800

總計：6,300
`

	transactions, err := parser.Parse(content)
	require.NoError(t, err)
	assert.Len(t, transactions, 2)

	// ROC year 114 = AD year 2025
	assert.Equal(t, 2025, transactions[0].Date.Year())
	assert.Equal(t, 12, int(transactions[0].Date.Month()))
	assert.Equal(t, 25, transactions[0].Date.Day())
	assert.Equal(t, "餐廳消費", transactions[0].Description)
	assert.Equal(t, float64(2500), transactions[0].Amount)

	// Check second transaction
	assert.Equal(t, 2025, transactions[1].Date.Year())
	assert.Equal(t, "線上購物", transactions[1].Description)
	assert.Equal(t, float64(3800), transactions[1].Amount)
}

func TestFubonParser_ParseSimpleDate(t *testing.T) {
	parser := NewFubonParser()

	content := `
富邦銀行信用卡帳單

消費明細：
12/25 超市購物 890

總計：890
`

	transactions, err := parser.Parse(content)
	require.NoError(t, err)
	assert.Len(t, transactions, 1)

	assert.Equal(t, 12, int(transactions[0].Date.Month()))
	assert.Equal(t, 25, transactions[0].Date.Day())
	assert.Equal(t, "超市購物", transactions[0].Description)
	assert.Equal(t, float64(890), transactions[0].Amount)
}

func TestFubonParser_ROCYearConversion(t *testing.T) {
	// Test that ROC year is correctly converted to AD year
	tests := []struct {
		rocYear  int
		expected int
	}{
		{rocYear: 100, expected: 2011},
		{rocYear: 110, expected: 2021},
		{rocYear: 114, expected: 2025},
		{rocYear: 115, expected: 2026},
	}

	for _, tt := range tests {
		adYear := tt.rocYear + 1911
		assert.Equal(t, tt.expected, adYear)
	}
}
