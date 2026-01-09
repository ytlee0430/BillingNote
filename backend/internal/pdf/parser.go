package pdf

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// Transaction represents a parsed transaction from PDF
type Transaction struct {
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Category    string    `json:"category,omitempty"`
	CardLast4   string    `json:"card_last4,omitempty"`
}

// BankParser interface for bank-specific parsers
type BankParser interface {
	// BankName returns the name of the bank
	BankName() string
	// CanParse checks if this parser can handle the given content
	CanParse(content string) bool
	// Parse extracts transactions from the PDF content
	Parse(content string) ([]Transaction, error)
}

// FileNameRule represents a rule for matching PDF files to banks
type FileNameRule struct {
	NameRule string `json:"name_rule"`
	Bank     string `json:"bank"`
	Password string `json:"password"`
}

// ParserRegistry manages bank parsers
type ParserRegistry struct {
	parsers       []BankParser
	fileNameRules []FileNameRule
}

// NewParserRegistry creates a new parser registry
func NewParserRegistry() *ParserRegistry {
	return &ParserRegistry{
		parsers:       make([]BankParser, 0),
		fileNameRules: make([]FileNameRule, 0),
	}
}

// RegisterParser adds a parser to the registry
func (r *ParserRegistry) RegisterParser(parser BankParser) {
	r.parsers = append(r.parsers, parser)
}

// LoadFileNameRules loads file name rules from JSON file
func (r *ParserRegistry) LoadFileNameRules(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file name rules: %w", err)
	}

	if err := json.Unmarshal(data, &r.fileNameRules); err != nil {
		return fmt.Errorf("failed to parse file name rules: %w", err)
	}

	return nil
}

// GetPasswordsForFile returns passwords to try for a given filename
func (r *ParserRegistry) GetPasswordsForFile(filename string) []string {
	passwords := make([]string, 0)

	for _, rule := range r.fileNameRules {
		// Convert glob-like pattern to regex
		pattern := globToRegex(rule.NameRule)
		matched, err := regexp.MatchString(pattern, filename)
		if err == nil && matched {
			passwords = append(passwords, rule.Password)
		}
	}

	return passwords
}

// globToRegex converts a glob pattern to a regex pattern
func globToRegex(glob string) string {
	// Escape special regex chars except * and ?
	result := regexp.QuoteMeta(glob)
	// Convert glob wildcards to regex
	result = strings.ReplaceAll(result, `\*`, `.*`)
	result = strings.ReplaceAll(result, `\?`, `.`)
	return "^" + result + "$"
}

// ExtractText extracts text from a PDF file with password attempts
func (r *ParserRegistry) ExtractText(pdfPath string, passwords []string) (string, error) {
	// Try without password first
	text, err := extractTextFromPDF(pdfPath, "")
	if err == nil {
		return text, nil
	}

	// Try each password
	for i, pwd := range passwords {
		text, err = extractTextFromPDF(pdfPath, pwd)
		if err == nil {
			fmt.Printf("PDF decrypted with password #%d\n", i+1)
			return text, nil
		}
	}

	return "", errors.New("all passwords failed or PDF is corrupted")
}

// extractTextFromPDF extracts text from a PDF file
func extractTextFromPDF(pdfPath string, password string) (string, error) {
	f, err := os.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	// Create configuration with password if provided
	conf := model.NewDefaultConfiguration()
	if password != "" {
		conf.UserPW = password
		conf.OwnerPW = password
	}

	// Read PDF into buffer
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, f); err != nil {
		return "", fmt.Errorf("failed to read PDF: %w", err)
	}

	// Extract text from all pages
	var textBuf bytes.Buffer
	if err := api.ExtractContent(bytes.NewReader(buf.Bytes()), &textBuf, nil, conf); err != nil {
		return "", fmt.Errorf("failed to extract text: %w", err)
	}

	return textBuf.String(), nil
}

// Parse parses a PDF file and returns transactions
func (r *ParserRegistry) Parse(pdfPath string, passwords []string) ([]Transaction, string, error) {
	// Extract text from PDF
	content, err := r.ExtractText(pdfPath, passwords)
	if err != nil {
		return nil, "", err
	}

	// Find matching parser
	for _, parser := range r.parsers {
		if parser.CanParse(content) {
			transactions, err := parser.Parse(content)
			if err != nil {
				return nil, parser.BankName(), fmt.Errorf("parser error: %w", err)
			}
			return transactions, parser.BankName(), nil
		}
	}

	return nil, "", errors.New("no suitable parser found for this PDF")
}

// ParseWithAutoPassword parses a PDF file using auto-detected passwords
func (r *ParserRegistry) ParseWithAutoPassword(pdfPath string) ([]Transaction, string, error) {
	// Get filename for password lookup
	filename := getFilename(pdfPath)
	passwords := r.GetPasswordsForFile(filename)

	return r.Parse(pdfPath, passwords)
}

func getFilename(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}
