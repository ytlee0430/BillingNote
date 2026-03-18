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

	// Real Cathay statement format (pdftotext -layout):
	//   消費日  入帳起息日  交易說明                              新臺幣金額  卡號後四碼  [行動卡號]  消費國家  幣別  [外幣金額]  [折算日]
	// Example:
	//   11/19   11/25   連加＊５０嵐（合江店）                      65    3842              TW   TWD
	//   11/21   11/24   LEETCODE.COM                            1,096    3842       9156   US   USD   35.00   11/20
	//
	// Primary pattern: uses card last-4 digits as anchor after amount
	// Amount is separated from description by 2+ spaces, followed by 4-digit card number
	withCardPattern := regexp.MustCompile(`^(\d{2}/\d{2})\s+\d{2}/\d{2}\s+(.+?)\s{2,}([\d,]+)\s+(\d{4})\b`)

	// Fallback pattern: for simpler formats without card number column
	fallbackPattern := regexp.MustCompile(`^(\d{2}/\d{2})\s+(?:\d{2}/\d{2}\s+)?(.+)\s+([\d,]+)\s*$`)

	// Lines to skip (summary/header lines that match date patterns but aren't transactions)
	skipKeywords := []string{
		"上期帳單總額", "繳款小計", "本行自動扣繳", "本期帳單總額",
		"新增消費小計", "循環利息", "最低應繳金額", "預借現金",
		"帳單結帳日", "繳款截止日", "page", "Page",
	}

	lines := strings.Split(content, "\n")
	currentYear := time.Now().Year()
	usedPrimary := false

	// First pass: try primary pattern (with card number anchor)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip known non-transaction lines
		skip := false
		for _, kw := range skipKeywords {
			if strings.Contains(line, kw) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		matches := withCardPattern.FindStringSubmatch(line)
		if len(matches) >= 5 {
			usedPrimary = true
			tx := p.parseTransaction(matches[1], matches[2], matches[3], matches[4], currentYear)
			if tx != nil {
				transactions = append(transactions, *tx)
			}
		}
	}

	// If primary pattern found transactions, return them
	if usedPrimary && len(transactions) > 0 {
		return transactions, nil
	}

	// Fallback: use simpler end-of-line anchored pattern
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		skip := false
		for _, kw := range skipKeywords {
			if strings.Contains(line, kw) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		matches := fallbackPattern.FindStringSubmatch(line)
		if len(matches) >= 4 {
			tx := p.parseTransaction(matches[1], matches[2], matches[3], "", currentYear)
			if tx != nil {
				transactions = append(transactions, *tx)
			}
		}
	}

	return transactions, nil
}

// parseTransaction creates a Transaction from parsed regex groups
func (p *CathayParser) parseTransaction(dateStr, description, amountStr, cardLast4 string, currentYear int) *pdf.Transaction {
	dateParts := strings.Split(dateStr, "/")
	if len(dateParts) != 2 {
		return nil
	}

	month, _ := strconv.Atoi(dateParts[0])
	day, _ := strconv.Atoi(dateParts[1])

	if month < 1 || month > 12 || day < 1 || day > 31 {
		return nil
	}

	// Determine year (handle year boundary)
	year := currentYear
	if month > int(time.Now().Month()) {
		year--
	}

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)

	description = strings.TrimSpace(description)
	if description == "" {
		return nil
	}

	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount == 0 {
		return nil
	}

	return &pdf.Transaction{
		Date:        date,
		Description: description,
		Amount:      amount,
		Currency:    "TWD",
		CardLast4:   cardLast4,
	}
}

var _ pdf.BankParser = (*CathayParser)(nil)
