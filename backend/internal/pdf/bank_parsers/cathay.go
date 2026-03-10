package bank_parsers

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"billing-note/internal/pdf"
)

// CathayParser parses 國泰世華 credit card statements
type CathayParser struct{}

// NewCathayParser creates a new Cathay parser
func NewCathayParser() *CathayParser {
	return &CathayParser{}
}

// BankName returns the bank name
func (p *CathayParser) BankName() string {
	return "國泰世華"
}

// CanParse checks if this parser can handle the content
func (p *CathayParser) CanParse(content string) bool {
	// Check for Cathay-specific keywords
	keywords := []string{"國泰世華", "CATHAY", "國泰銀行"}
	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

// Parse extracts transactions from Cathay credit card statement
func (p *CathayParser) Parse(content string) ([]pdf.Transaction, error) {
	transactions := make([]pdf.Transaction, 0)

	// Cathay statement format: MM/DD  [MM/DD]  description  amount
	// The amount is always the LAST number on the line.
	// Use greedy .+ so amount captures only the trailing number.
	// Optionally match a second MM/DD (posting date) right after the first.
	datePattern := regexp.MustCompile(`^(\d{2}/\d{2})\s+(?:\d{2}/\d{2}\s+)?(.+)\s+([\d,]+)\s*$`)

	lines := strings.Split(content, "\n")
	currentYear := time.Now().Year()

	for _, line := range lines {
		line = strings.TrimSpace(line)
		matches := datePattern.FindStringSubmatch(line)

		if len(matches) >= 4 {
			// Parse date (MM/DD)
			dateParts := strings.Split(matches[1], "/")
			if len(dateParts) == 2 {
				month, _ := strconv.Atoi(dateParts[0])
				day, _ := strconv.Atoi(dateParts[1])

				if month < 1 || month > 12 || day < 1 || day > 31 {
					continue
				}

				// Determine year (handle year boundary)
				year := currentYear
				if month > int(time.Now().Month()) {
					year--
				}

				date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)

				// Parse description (trim trailing spaces from greedy match)
				description := strings.TrimSpace(matches[2])
				if description == "" {
					continue
				}

				// Parse amount (remove commas)
				amountStr := strings.ReplaceAll(matches[3], ",", "")
				amount, err := strconv.ParseFloat(amountStr, 64)
				if err != nil || amount == 0 {
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
	}

	return transactions, nil
}

var _ pdf.BankParser = (*CathayParser)(nil)
