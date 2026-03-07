package services

import (
	"billing-note/internal/models"
	"billing-note/internal/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Transaction Repository ---

type mockTransactionRepo struct {
	mock.Mock
}

func (m *mockTransactionRepo) Create(transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *mockTransactionRepo) GetByID(id uint) (*models.Transaction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *mockTransactionRepo) List(filter repository.TransactionFilter) ([]models.Transaction, int64, error) {
	args := m.Called(filter)
	return args.Get(0).([]models.Transaction), args.Get(1).(int64), args.Error(2)
}

func (m *mockTransactionRepo) Update(transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *mockTransactionRepo) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockTransactionRepo) GetMonthlyStats(userID uint, year int, month int) (map[string]float64, error) {
	args := m.Called(userID, year, month)
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *mockTransactionRepo) GetCategoryStats(userID uint, startDate, endDate time.Time, transactionType string) ([]map[string]interface{}, error) {
	args := m.Called(userID, startDate, endDate, transactionType)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

// --- Helper ---

func newTestDeduplicationService(txnRepo *mockTransactionRepo, invRepo *mockInvoiceRepo) *DeduplicationService {
	return NewDeduplicationService(txnRepo, invRepo)
}

func makeInvoice(userID uint, number string, amount float64, date time.Time, seller string) *models.Invoice {
	return &models.Invoice{
		ID:            1,
		UserID:        userID,
		InvoiceNumber: number,
		Amount:        amount,
		InvoiceDate:   date,
		SellerName:    seller,
	}
}

func makeTransaction(id uint, description string, amount float64, date time.Time) models.Transaction {
	return models.Transaction{
		ID:              id,
		Amount:          amount,
		Description:     description,
		TransactionDate: date,
	}
}

// --- Levenshtein Tests ---

func TestLevenshteinSimilarity_Identical(t *testing.T) {
	assert.Equal(t, 1.0, LevenshteinSimilarity("hello", "hello"))
}

func TestLevenshteinSimilarity_Empty(t *testing.T) {
	assert.Equal(t, 1.0, LevenshteinSimilarity("", ""))
	assert.Equal(t, 0.0, LevenshteinSimilarity("hello", ""))
	assert.Equal(t, 0.0, LevenshteinSimilarity("", "hello"))
}

func TestLevenshteinSimilarity_Similar(t *testing.T) {
	// "全家便利商店" vs "全家" -> distance=4, maxLen=6, similarity=0.333...
	sim := LevenshteinSimilarity("全家便利商店", "全家")
	assert.InDelta(t, 0.333, sim, 0.01)
}

func TestLevenshteinSimilarity_CompletelyDifferent(t *testing.T) {
	sim := LevenshteinSimilarity("abc", "xyz")
	assert.Equal(t, 0.0, sim)
}

func TestLevenshteinSimilarity_OneCharDiff(t *testing.T) {
	// "hello" vs "hallo" -> distance=1, maxLen=5, similarity=0.8
	sim := LevenshteinSimilarity("hello", "hallo")
	assert.InDelta(t, 0.8, sim, 0.001)
}

// --- FindDuplicates Tests ---

func TestDeduplication_ExactMatch(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	invRepo := new(mockInvoiceRepo)
	svc := newTestDeduplicationService(txnRepo, invRepo)

	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	invoice := makeInvoice(1, "AB12345678", 150.0, invoiceDate, "全家便利商店")

	transactions := []models.Transaction{
		makeTransaction(10, "全家便利商店", 150.0, invoiceDate),
	}

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return(transactions, int64(1), nil)

	matches := svc.FindDuplicates(invoice)

	assert.Len(t, matches, 1)
	assert.Equal(t, uint(10), matches[0].Transaction.ID)
	assert.InDelta(t, 1.0, matches[0].ConfidenceScore, 0.05)
}

func TestDeduplication_SameAmountDifferentDate(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	invRepo := new(mockInvoiceRepo)
	svc := newTestDeduplicationService(txnRepo, invRepo)

	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	txnDate := time.Date(2026, 1, 17, 0, 0, 0, 0, time.UTC) // 2 days later
	invoice := makeInvoice(1, "AB12345678", 200.0, invoiceDate, "7-ELEVEN")

	transactions := []models.Transaction{
		makeTransaction(20, "7-ELEVEN", 200.0, txnDate),
	}

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return(transactions, int64(1), nil)

	matches := svc.FindDuplicates(invoice)

	assert.Len(t, matches, 1)
	// Should match but with lower confidence due to date difference
	assert.Less(t, matches[0].ConfidenceScore, 1.0)
	assert.Greater(t, matches[0].ConfidenceScore, 0.7)
}

func TestDeduplication_SameDate_DifferentAmount(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	invRepo := new(mockInvoiceRepo)
	svc := newTestDeduplicationService(txnRepo, invRepo)

	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	invoice := makeInvoice(1, "AB12345678", 100.0, invoiceDate, "全家便利商店")

	transactions := []models.Transaction{
		// Amount diff = 0.5, within tolerance of 1.0
		makeTransaction(30, "全家便利商店", 100.5, invoiceDate),
	}

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return(transactions, int64(1), nil)

	matches := svc.FindDuplicates(invoice)

	assert.Len(t, matches, 1)
	assert.Greater(t, matches[0].ConfidenceScore, 0.9)
}

func TestDeduplication_AmountExceedsTolerance(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	invRepo := new(mockInvoiceRepo)
	svc := newTestDeduplicationService(txnRepo, invRepo)

	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	invoice := makeInvoice(1, "AB12345678", 100.0, invoiceDate, "全家便利商店")

	transactions := []models.Transaction{
		// Amount diff = 2.0, exceeds tolerance of 1.0
		makeTransaction(30, "全家便利商店", 102.0, invoiceDate),
	}

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return(transactions, int64(1), nil)

	matches := svc.FindDuplicates(invoice)

	assert.Len(t, matches, 0)
}

func TestDeduplication_SimilarMerchant_80percent(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	invRepo := new(mockInvoiceRepo)
	svc := newTestDeduplicationService(txnRepo, invRepo)

	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	// "hello" vs "hallo" = 0.8 similarity (exactly at threshold)
	invoice := makeInvoice(1, "AB12345678", 100.0, invoiceDate, "hello")

	transactions := []models.Transaction{
		makeTransaction(40, "hallo", 100.0, invoiceDate),
	}

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return(transactions, int64(1), nil)

	matches := svc.FindDuplicates(invoice)

	assert.Len(t, matches, 1)
}

func TestDeduplication_SimilarMerchant_79percent(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	invRepo := new(mockInvoiceRepo)
	svc := newTestDeduplicationService(txnRepo, invRepo)

	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	// "abcde" vs "abxyz" = distance 3, maxLen 5, similarity 0.4 (below 0.8)
	invoice := makeInvoice(1, "AB12345678", 100.0, invoiceDate, "abcde")

	transactions := []models.Transaction{
		makeTransaction(50, "abxyz", 100.0, invoiceDate),
	}

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return(transactions, int64(1), nil)

	matches := svc.FindDuplicates(invoice)

	assert.Len(t, matches, 0)
}

func TestDeduplication_MultipleMatches(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	invRepo := new(mockInvoiceRepo)
	svc := newTestDeduplicationService(txnRepo, invRepo)

	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	invoice := makeInvoice(1, "AB12345678", 100.0, invoiceDate, "全家便利商店")

	transactions := []models.Transaction{
		makeTransaction(60, "全家便利商店", 100.0, invoiceDate),                                                   // exact match
		makeTransaction(61, "全家便利商店", 100.5, invoiceDate),                                                   // slight amount diff
		makeTransaction(62, "全家便利商店", 100.0, time.Date(2026, 1, 17, 0, 0, 0, 0, time.UTC)), // 2 days later
	}

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return(transactions, int64(3), nil)

	matches := svc.FindDuplicates(invoice)

	assert.Len(t, matches, 3)
	// First match should have highest confidence (exact match)
	assert.GreaterOrEqual(t, matches[0].ConfidenceScore, matches[1].ConfidenceScore)
	assert.GreaterOrEqual(t, matches[1].ConfidenceScore, matches[2].ConfidenceScore)
}

func TestDeduplication_EdgeCase_Timezone(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	invRepo := new(mockInvoiceRepo)
	svc := newTestDeduplicationService(txnRepo, invRepo)

	// Invoice at end of day UTC
	invoiceDate := time.Date(2026, 1, 15, 23, 59, 0, 0, time.UTC)
	// Transaction at start of next day UTC
	txnDate := time.Date(2026, 1, 16, 0, 1, 0, 0, time.UTC)

	invoice := makeInvoice(1, "AB12345678", 100.0, invoiceDate, "7-ELEVEN")

	transactions := []models.Transaction{
		makeTransaction(70, "7-ELEVEN", 100.0, txnDate),
	}

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return(transactions, int64(1), nil)

	matches := svc.FindDuplicates(invoice)

	// Should still match - within the 3-day tolerance
	assert.Len(t, matches, 1)
	assert.Greater(t, matches[0].ConfidenceScore, 0.8)
}

func TestDeduplication_NoTransactions(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	invRepo := new(mockInvoiceRepo)
	svc := newTestDeduplicationService(txnRepo, invRepo)

	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	invoice := makeInvoice(1, "AB12345678", 100.0, invoiceDate, "全家")

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return([]models.Transaction{}, int64(0), nil)

	matches := svc.FindDuplicates(invoice)

	assert.Len(t, matches, 0)
}

func TestDeduplication_TransactionRepoError(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	invRepo := new(mockInvoiceRepo)
	svc := newTestDeduplicationService(txnRepo, invRepo)

	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	invoice := makeInvoice(1, "AB12345678", 100.0, invoiceDate, "全家")

	txnRepo.On("List", mock.AnythingOfType("repository.TransactionFilter")).
		Return([]models.Transaction{}, int64(0), assert.AnError)

	matches := svc.FindDuplicates(invoice)

	assert.Nil(t, matches)
}

// --- RunDeduplication Tests ---

func TestRunDeduplication_MatchesFound(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	invRepo := new(mockInvoiceRepo)
	svc := newTestDeduplicationService(txnRepo, invRepo)

	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	invoices := []models.Invoice{
		{ID: 1, UserID: 1, InvoiceNumber: "AB11111111", Amount: 100, InvoiceDate: invoiceDate, SellerName: "全家便利商店", IsDuplicated: false},
		{ID: 2, UserID: 1, InvoiceNumber: "AB22222222", Amount: 200, InvoiceDate: invoiceDate, SellerName: "7-ELEVEN", IsDuplicated: false},
		{ID: 3, UserID: 1, InvoiceNumber: "AB33333333", Amount: 300, InvoiceDate: invoiceDate, SellerName: "萊爾富", IsDuplicated: true}, // already duplicated
	}

	invRepo.On("List", uint(1), (*time.Time)(nil), (*time.Time)(nil), 0, 0).
		Return(invoices, int64(3), nil)

	// For invoice 1: match found
	txnRepo.On("List", mock.MatchedBy(func(f repository.TransactionFilter) bool {
		return f.UserID == 1
	})).Return([]models.Transaction{
		makeTransaction(10, "全家便利商店", 100, invoiceDate),
	}, int64(1), nil).Once()

	// For invoice 2: no match (different merchant name)
	txnRepo.On("List", mock.MatchedBy(func(f repository.TransactionFilter) bool {
		return f.UserID == 1
	})).Return([]models.Transaction{}, int64(0), nil).Once()

	invRepo.On("Update", mock.AnythingOfType("*models.Invoice")).Return(nil)

	matched, err := svc.RunDeduplication(1)

	assert.NoError(t, err)
	assert.Equal(t, 1, matched)
	invRepo.AssertNumberOfCalls(t, "Update", 1)
}

func TestRunDeduplication_SkipsAlreadyDuplicated(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	invRepo := new(mockInvoiceRepo)
	svc := newTestDeduplicationService(txnRepo, invRepo)

	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	invoices := []models.Invoice{
		{ID: 1, UserID: 1, InvoiceNumber: "AB11111111", Amount: 100, InvoiceDate: invoiceDate, SellerName: "全家", IsDuplicated: true},
	}

	invRepo.On("List", uint(1), (*time.Time)(nil), (*time.Time)(nil), 0, 0).
		Return(invoices, int64(1), nil)

	matched, err := svc.RunDeduplication(1)

	assert.NoError(t, err)
	assert.Equal(t, 0, matched)
	// Should never call transaction List since all invoices are already duplicated
	txnRepo.AssertNotCalled(t, "List", mock.Anything)
}

func TestRunDeduplication_InvoiceRepoError(t *testing.T) {
	txnRepo := new(mockTransactionRepo)
	invRepo := new(mockInvoiceRepo)
	svc := newTestDeduplicationService(txnRepo, invRepo)

	invRepo.On("List", uint(1), (*time.Time)(nil), (*time.Time)(nil), 0, 0).
		Return([]models.Invoice{}, int64(0), assert.AnError)

	matched, err := svc.RunDeduplication(1)

	assert.Error(t, err)
	assert.Equal(t, 0, matched)
}

// --- calculateConfidence Tests ---

func TestCalculateConfidence_ExactMatch(t *testing.T) {
	svc := NewDeduplicationService(nil, nil)

	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	confidence := svc.calculateConfidence(0.0, 1.0, date, date)

	assert.InDelta(t, 1.0, confidence, 0.01)
}

func TestCalculateConfidence_WorstCase(t *testing.T) {
	svc := NewDeduplicationService(nil, nil)

	txnDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	invDate := time.Date(2026, 1, 18, 0, 0, 0, 0, time.UTC) // 3 days diff

	// amountDiff=1.0 (max tolerance), similarity=0.8 (min threshold)
	confidence := svc.calculateConfidence(1.0, 0.8, txnDate, invDate)

	assert.Greater(t, confidence, 0.5)
	assert.Less(t, confidence, 1.0)
}
