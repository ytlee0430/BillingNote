package bank_parsers

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"billing-note/internal/pdf"
)

// TaishinParser parses 台新銀行 credit card statements
type TaishinParser struct{}

// NewTaishinParser creates a new Taishin parser
func NewTaishinParser() *TaishinParser {
	return &TaishinParser{}
}

// BankName returns the bank name
func (p *TaishinParser) BankName() string {
	return "台新銀行"
}

// CanParse checks if this parser can handle the content
func (p *TaishinParser) CanParse(content string) bool {
	// Check for Taishin-specific keywords
	keywords := []string{"台新銀行", "台新國際商業銀行", "TAISHIN", "TSB"}
	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

// Parse extracts transactions from Taishin credit card statement
func (p *TaishinParser) Parse(content string) ([]pdf.Transaction, error) {
	transactions := make([]pdf.Transaction, 0)

	// Taishin statement patterns
	// Format varies: YYYY/MM/DD or MM/DD description amount
	datePattern := regexp.MustCompile(`(\d{4}/\d{2}/\d{2}|\d{2}/\d{2})\s+(.+?)\s+([\d,]+)`)

	lines := strings.Split(content, "\n")
	currentYear := time.Now().Year()

	for _, line := range lines {
		line = strings.TrimSpace(line)
		matches := datePattern.FindStringSubmatch(line)

		if len(matches) >= 4 {
			dateStr := matches[1]
			var date time.Time

			if strings.Count(dateStr, "/") == 2 {
				// YYYY/MM/DD format
				parts := strings.Split(dateStr, "/")
				year, _ := strconv.Atoi(parts[0])
				month, _ := strconv.Atoi(parts[1])
				day, _ := strconv.Atoi(parts[2])
				date = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
			} else {
				// MM/DD format
				parts := strings.Split(dateStr, "/")
				month, _ := strconv.Atoi(parts[0])
				day, _ := strconv.Atoi(parts[1])
				year := currentYear
				if month > int(time.Now().Month()) {
					year--
				}
				date = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
			}

			description := strings.TrimSpace(matches[2])
			amountStr := strings.ReplaceAll(matches[3], ",", "")
			amount, err := strconv.ParseFloat(amountStr, 64)
			if err != nil {
				continue
			}

			transactions = append(transactions, pdf.Transaction{
				Date:        date,
				Description: description,
				Amount:      amount,
				Currency:    "TWD",
			})
		}
	}

	return transactions, nil
}

var _ pdf.BankParser = (*TaishinParser)(nil)
