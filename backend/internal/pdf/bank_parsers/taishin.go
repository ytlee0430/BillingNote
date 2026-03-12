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
//
// Taishin pdftotext -layout output has two transaction formats:
//
// Type A (single-line): date + posting date + description + amount + TW
//   114/11/26 114/11/28     街口電支－臺北自來水事業處TAIPEI              551     TW
//
// Type B (multi-line): description on indented line(s) above, date + amount on date line
//                           街口電支－大台北區瓦斯股份有限
//   114/12/06 114/12/08                                      792     TW
//                           TAIPEI
//
// Card sections are delimited by: (卡號末四碼:XXXX)
func (p *TaishinParser) Parse(content string) ([]pdf.Transaction, error) {
	transactions := make([]pdf.Transaction, 0)
	lines := strings.Split(content, "\n")

	// Card section header: extract last 4 digits
	cardHeaderRe := regexp.MustCompile(`卡號末四碼[:：](\d{4})`)

	// Type A: single-line with inline description
	singleLineRe := regexp.MustCompile(`^(\d{3}/\d{2}/\d{2})\s+\d{3}/\d{2}/\d{2}\s+(.+?)\s{2,}(-?[\d,]+)\s+TW`)

	// Type B: date line with amount only (description on adjacent indented lines)
	amountOnlyRe := regexp.MustCompile(`^(\d{3}/\d{2}/\d{2})\s+\d{3}/\d{2}/\d{2}\s+(-?[\d,]+)(?:\s+TW)?\s*$`)

	// Indented text line (10+ leading spaces then non-space content)
	indentedRe := regexp.MustCompile(`^\s{10,}(\S.*)$`)

	// Skip keywords for non-transaction indented lines
	skipKeywords := []string{"卡號末四碼", "消費日", "入帳起息日", "新臺幣金額", "外幣", "帳務資訊"}

	currentCard := ""

	for i, line := range lines {
		// Check for card section header
		if m := cardHeaderRe.FindStringSubmatch(line); len(m) >= 2 {
			currentCard = m[1]
			continue
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Try Type B first (amount-only date line) — more specific, avoids false Type A match
		if m := amountOnlyRe.FindStringSubmatch(trimmed); len(m) >= 3 {
			desc := p.collectDescription(lines, i, indentedRe, skipKeywords)
			tx := p.buildTransaction(m[1], desc, m[2], currentCard)
			if tx != nil {
				transactions = append(transactions, *tx)
			}
			continue
		}

		// Try Type A (single-line with inline description)
		if m := singleLineRe.FindStringSubmatch(trimmed); len(m) >= 4 {
			tx := p.buildTransaction(m[1], strings.TrimSpace(m[2]), m[3], currentCard)
			if tx != nil {
				transactions = append(transactions, *tx)
			}
		}
	}

	return transactions, nil
}

// collectDescription scans backward from a date line to find indented description lines.
// Stops at date lines, card headers, skip keywords, or non-indented lines.
// Only collects the FIRST indented line directly above the date line (skipping
// continuation lines from previous transactions like "TAIPEI" or "/TW").
func (p *TaishinParser) collectDescription(lines []string, dateLineIdx int, indentedRe *regexp.Regexp, skipKws []string) string {
	dateLineRe := regexp.MustCompile(`^\d{3}/\d{2}/\d{2}`)

	// Scan backward, collecting all consecutive indented lines
	allParts := []string{}
	for j := dateLineIdx - 1; j >= 0; j-- {
		line := lines[j]
		trimmed := strings.TrimSpace(line)

		// Stop at empty lines
		if trimmed == "" {
			break
		}

		// Stop at date lines (previous transaction)
		if dateLineRe.MatchString(trimmed) {
			break
		}

		m := indentedRe.FindStringSubmatch(line)
		if m == nil {
			break
		}

		text := strings.TrimSpace(m[1])
		skip := false
		for _, kw := range skipKws {
			if strings.Contains(text, kw) {
				skip = true
				break
			}
		}
		if skip {
			break
		}
		allParts = append([]string{text}, allParts...)
	}

	// If multiple indented lines collected and the first one looks like a
	// continuation from a previous tx (short location-only like "TAIPEI", "/TW"),
	// skip it and only use the actual description line(s).
	if len(allParts) > 1 {
		first := allParts[0]
		// Continuation lines are typically short location suffixes
		if len(first) <= 10 && !strings.ContainsAny(first, "－＊＝") {
			allParts = allParts[1:]
		}
	}

	return strings.Join(allParts, "")
}

// buildTransaction creates a Transaction from parsed ROC date and amount strings
func (p *TaishinParser) buildTransaction(rocDateStr, description, amountStr, cardLast4 string) *pdf.Transaction {
	// Parse ROC date: 114/12/02 → 2025/12/02
	parts := strings.Split(rocDateStr, "/")
	if len(parts) != 3 {
		return nil
	}

	rocYear, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])
	year := rocYear + 1911

	if month < 1 || month > 12 || day < 1 || day > 31 {
		return nil
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

var _ pdf.BankParser = (*TaishinParser)(nil)
