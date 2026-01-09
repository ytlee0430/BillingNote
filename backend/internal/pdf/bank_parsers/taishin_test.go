package bank_parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaishinParser_BankName(t *testing.T) {
	parser := NewTaishinParser()
	assert.Equal(t, "台新銀行", parser.BankName())
}

func TestTaishinParser_CanParse(t *testing.T) {
	parser := NewTaishinParser()

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "contains 台新銀行",
			content: "這是台新銀行的帳單",
			want:    true,
		},
		{
			name:    "contains TAISHIN",
			content: "TAISHIN BANK Statement",
			want:    true,
		},
		{
			name:    "contains TSB",
			content: "TSB Credit Card Statement",
			want:    true,
		},
		{
			name:    "contains 台新國際商業銀行",
			content: "台新國際商業銀行信用卡帳單",
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

func TestTaishinParser_Parse(t *testing.T) {
	parser := NewTaishinParser()

	content := `
台新銀行信用卡帳單

消費明細：
2025/12/25 百貨公司購物 5,000
2025/12/26 加油站 1,200
12/27 便利商店 120

總計：6,320
`

	transactions, err := parser.Parse(content)
	require.NoError(t, err)
	assert.Len(t, transactions, 3)

	// Check full date format transaction
	assert.Equal(t, 2025, transactions[0].Date.Year())
	assert.Equal(t, 12, int(transactions[0].Date.Month()))
	assert.Equal(t, 25, transactions[0].Date.Day())
	assert.Equal(t, "百貨公司購物", transactions[0].Description)
	assert.Equal(t, float64(5000), transactions[0].Amount)

	// Check second transaction
	assert.Equal(t, "加油站", transactions[1].Description)
	assert.Equal(t, float64(1200), transactions[1].Amount)

	// Check short date format transaction
	assert.Equal(t, "便利商店", transactions[2].Description)
	assert.Equal(t, float64(120), transactions[2].Amount)
}
