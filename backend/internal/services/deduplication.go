package services

import (
	"billing-note/internal/models"
	"billing-note/internal/repository"
	"billing-note/pkg/logger"
	"math"
	"time"
	"unicode/utf8"
)

// DuplicateMatch represents a potential duplicate match
type DuplicateMatch struct {
	Transaction     models.Transaction `json:"transaction"`
	ConfidenceScore float64            `json:"confidence_score"`
}

// DeduplicationService handles invoice-transaction deduplication
type DeduplicationService struct {
	transactionRepo repository.TransactionRepository
	invoiceRepo     repository.InvoiceRepository
	amountTolerance float64
	daysTolerance   int
	minSimilarity   float64
}

// NewDeduplicationService creates a new deduplication service
func NewDeduplicationService(
	transactionRepo repository.TransactionRepository,
	invoiceRepo repository.InvoiceRepository,
) *DeduplicationService {
	return &DeduplicationService{
		transactionRepo: transactionRepo,
		invoiceRepo:     invoiceRepo,
		amountTolerance: 1.0,
		daysTolerance:   3,
		minSimilarity:   0.8,
	}
}

// FindDuplicates finds potential transaction matches for an invoice
func (s *DeduplicationService) FindDuplicates(invoice *models.Invoice) []DuplicateMatch {
	log := logger.ServiceLog("DeduplicationService", "FindDuplicates")

	startDate := invoice.InvoiceDate.AddDate(0, 0, -s.daysTolerance)
	endDate := invoice.InvoiceDate.AddDate(0, 0, s.daysTolerance)

	// Find transactions in date range
	filter := repository.TransactionFilter{
		UserID:    invoice.UserID,
		StartDate: &startDate,
		EndDate:   &endDate,
		Page:      1,
		PageSize:  100,
	}

	transactions, _, err := s.transactionRepo.List(filter)
	if err != nil {
		log.WithError(err).Warn("Failed to query transactions for dedup")
		return nil
	}

	var matches []DuplicateMatch

	for _, txn := range transactions {
		// Check amount tolerance
		amountDiff := math.Abs(txn.Amount - invoice.Amount)
		if amountDiff > s.amountTolerance {
			continue
		}

		// Calculate merchant name similarity
		similarity := LevenshteinSimilarity(txn.Description, invoice.SellerName)
		if similarity < s.minSimilarity {
			continue
		}

		// Calculate confidence score
		confidence := s.calculateConfidence(amountDiff, similarity, txn.TransactionDate, invoice.InvoiceDate)

		matches = append(matches, DuplicateMatch{
			Transaction:     txn,
			ConfidenceScore: confidence,
		})
	}

	// Sort by confidence (highest first)
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].ConfidenceScore > matches[i].ConfidenceScore {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	log.WithFields(logger.Fields{
		"invoice_number": invoice.InvoiceNumber,
		"matches_found":  len(matches),
	}).Debug("Deduplication check completed")

	return matches
}

// RunDeduplication runs deduplication for all unmatched invoices of a user
func (s *DeduplicationService) RunDeduplication(userID uint) (int, error) {
	log := logger.ServiceLog("DeduplicationService", "RunDeduplication")

	// Get all non-duplicated invoices
	invoices, _, err := s.invoiceRepo.List(userID, nil, nil, 0, 0)
	if err != nil {
		return 0, err
	}

	matched := 0
	for i := range invoices {
		if invoices[i].IsDuplicated {
			continue
		}

		matches := s.FindDuplicates(&invoices[i])
		if len(matches) > 0 {
			best := matches[0]
			invoices[i].IsDuplicated = true
			invoices[i].DuplicatedTransactionID = &best.Transaction.ID
			invoices[i].ConfidenceScore = &best.ConfidenceScore

			if err := s.invoiceRepo.Update(&invoices[i]); err != nil {
				log.WithError(err).Warn("Failed to update invoice with duplicate info")
				continue
			}
			matched++
		}
	}

	log.WithFields(logger.Fields{
		"user_id":        userID,
		"total_invoices": len(invoices),
		"matched":        matched,
	}).Info("Deduplication run completed")

	return matched, nil
}

// calculateConfidence computes a confidence score (0.0 - 1.0)
func (s *DeduplicationService) calculateConfidence(amountDiff, similarity float64, txnDate, invDate time.Time) float64 {
	// Amount score: exact match = 1.0, max tolerance = 0.8
	amountScore := 1.0 - (amountDiff / (s.amountTolerance + 1.0)) * 0.2

	// Date score: same day = 1.0, 3 days away = 0.7
	daysDiff := math.Abs(txnDate.Sub(invDate).Hours() / 24)
	dateScore := 1.0 - (daysDiff / float64(s.daysTolerance+1)) * 0.3

	// Similarity score directly
	similarityScore := similarity

	// Weighted average
	confidence := amountScore*0.3 + dateScore*0.3 + similarityScore*0.4

	// Clamp to [0, 1]
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return math.Round(confidence*100) / 100
}

// LevenshteinSimilarity calculates the similarity between two strings using Levenshtein distance
func LevenshteinSimilarity(a, b string) float64 {
	if a == "" && b == "" {
		return 1.0
	}
	if a == "" || b == "" {
		return 0.0
	}

	distance := levenshteinDistance(a, b)
	maxLen := utf8.RuneCountInString(a)
	bLen := utf8.RuneCountInString(b)
	if bLen > maxLen {
		maxLen = bLen
	}

	return 1.0 - float64(distance)/float64(maxLen)
}

// levenshteinDistance computes the Levenshtein distance between two strings
func levenshteinDistance(a, b string) int {
	runesA := []rune(a)
	runesB := []rune(b)
	lenA := len(runesA)
	lenB := len(runesB)

	if lenA == 0 {
		return lenB
	}
	if lenB == 0 {
		return lenA
	}

	// Create matrix
	matrix := make([][]int, lenA+1)
	for i := range matrix {
		matrix[i] = make([]int, lenB+1)
		matrix[i][0] = i
	}
	for j := 0; j <= lenB; j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= lenA; i++ {
		for j := 1; j <= lenB; j++ {
			cost := 1
			if runesA[i-1] == runesB[j-1] {
				cost = 0
			}
			matrix[i][j] = min3(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[lenA][lenB]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
