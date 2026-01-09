package bank_parsers

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"billing-note/internal/pdf"
)

// FubonParser parses 富邦銀行 credit card statements
type FubonParser struct{}

// NewFubonParser creates a new Fubon parser
func NewFubonParser() *FubonParser {
	return &FubonParser{}
}

// BankName returns the bank name
func (p *FubonParser) BankName() string {
	return "富邦銀行"
}

// CanParse checks if this parser can handle the content
func (p *FubonParser) CanParse(content string) bool {
	// Check for Fubon-specific keywords
	keywords := []string{"富邦銀行", "台北富邦", "FUBON", "富邦金控"}
	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

// Parse extracts transactions from Fubon credit card statement
func (p *FubonParser) Parse(content string) ([]pdf.Transaction, error) {
	transactions := make([]pdf.Transaction, 0)

	// Fubon uses ROC calendar (民國年)
	// Format: 114/12/25 or 12/25 description amount
	rocDatePattern := regexp.MustCompile(`(\d{3}/\d{2}/\d{2})\s+(.+?)\s+([\d,]+)`)
	simpleDatePattern := regexp.MustCompile(`(\d{2}/\d{2})\s+(.+?)\s+([\d,]+)`)

	lines := strings.Split(content, "\n")
	currentYear := time.Now().Year()

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Try ROC date format first (114/12/25)
		matches := rocDatePattern.FindStringSubmatch(line)
		if len(matches) >= 4 {
			parts := strings.Split(matches[1], "/")
			rocYear, _ := strconv.Atoi(parts[0])
			month, _ := strconv.Atoi(parts[1])
			day, _ := strconv.Atoi(parts[2])

			// Convert ROC year to AD year (民國 + 1911 = 西元)
			year := rocYear + 1911
			date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)

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
			continue
		}

		// Try simple date format (12/25)
		matches = simpleDatePattern.FindStringSubmatch(line)
		if len(matches) >= 4 {
			parts := strings.Split(matches[1], "/")
			month, _ := strconv.Atoi(parts[0])
			day, _ := strconv.Atoi(parts[1])

			year := currentYear
			if month > int(time.Now().Month()) {
				year--
			}
			date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)

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

var _ pdf.BankParser = (*FubonParser)(nil)
