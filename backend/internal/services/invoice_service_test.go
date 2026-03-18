package services

import (
	"billing-note/internal/models"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock MOF API Client ---

type mockMOFClient struct {
	mock.Mock
}

func (m *mockMOFClient) FetchInvoices(carrierCode, startDate, endDate string) (*MOFResponse, error) {
	args := m.Called(carrierCode, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*MOFResponse), args.Error(1)
}

// --- Mock Invoice Repository ---

type mockInvoiceRepo struct {
	mock.Mock
}

func (m *mockInvoiceRepo) Create(invoice *models.Invoice) error {
	args := m.Called(invoice)
	return args.Error(0)
}

func (m *mockInvoiceRepo) GetByID(id uint) (*models.Invoice, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invoice), args.Error(1)
}

func (m *mockInvoiceRepo) GetByInvoiceNumber(userID uint, invoiceNumber string) (*models.Invoice, error) {
	args := m.Called(userID, invoiceNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invoice), args.Error(1)
}

func (m *mockInvoiceRepo) List(userID uint, startDate, endDate *time.Time, page, pageSize int) ([]models.Invoice, int64, error) {
	args := m.Called(userID, startDate, endDate, page, pageSize)
	return args.Get(0).([]models.Invoice), args.Get(1).(int64), args.Error(2)
}

func (m *mockInvoiceRepo) Update(invoice *models.Invoice) error {
	args := m.Called(invoice)
	return args.Error(0)
}

func (m *mockInvoiceRepo) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockInvoiceRepo) BatchCreate(invoices []*models.Invoice) (int, error) {
	args := m.Called(invoices)
	return args.Int(0), args.Error(1)
}

// --- Tests ---

func newTestInvoiceService(repo *mockInvoiceRepo, mofClient *mockMOFClient) *InvoiceService {
	svc := NewInvoiceService(repo, "https://api.test.com", "test-app-id")
	if mofClient != nil {
		svc.SetMOFClient(mofClient)
	}
	return svc
}

func TestSyncInvoices_Success(t *testing.T) {
	repo := new(mockInvoiceRepo)
	mofClient := new(mockMOFClient)
	svc := newTestInvoiceService(repo, mofClient)

	mofResp := &MOFResponse{
		Version: "0.5",
		Code:    200,
		Message: "成功",
		Details: []MOFInvoice{
			{
				InvNum:     "AB12345678",
				SellerName: "全家便利商店",
				Amount:     json.Number("150"),
				InvDate:    "2026/01/05 14:30:00",
				SellerBAN:  "12345678",
				InvStatus:  "已使用",
			},
			{
				InvNum:     "CD87654321",
				SellerName: "7-ELEVEN",
				Amount:     json.Number("89"),
				InvDate:    "2026/01/06 10:00:00",
				SellerBAN:  "87654321",
				InvStatus:  "已使用",
			},
		},
	}

	mofClient.On("FetchInvoices", "/ABCD123", "2026/01/01", "2026/01/31").Return(mofResp, nil)
	repo.On("BatchCreate", mock.AnythingOfType("[]*models.Invoice")).Return(2, nil)

	created, err := svc.SyncInvoices(1, "/ABCD123", "2026/01/01", "2026/01/31")
	assert.NoError(t, err)
	assert.Equal(t, 2, created)

	// Verify the invoices were created with correct data
	invoices := repo.Calls[0].Arguments.Get(0).([]*models.Invoice)
	assert.Len(t, invoices, 2)
	assert.Equal(t, "AB12345678", invoices[0].InvoiceNumber)
	assert.Equal(t, "全家便利商店", invoices[0].SellerName)
	assert.Equal(t, float64(150), invoices[0].Amount)
	assert.Equal(t, uint(1), invoices[0].UserID)
}

func TestSyncInvoices_EmptyCarrier(t *testing.T) {
	repo := new(mockInvoiceRepo)
	svc := newTestInvoiceService(repo, nil)

	_, err := svc.SyncInvoices(1, "", "2026/01/01", "2026/01/31")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "carrier code is required")
}

func TestSyncInvoices_InvalidCarrierFormat(t *testing.T) {
	repo := new(mockInvoiceRepo)
	svc := newTestInvoiceService(repo, nil)

	_, err := svc.SyncInvoices(1, "ABCD123", "2026/01/01", "2026/01/31")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid carrier code")
}

func TestSyncInvoices_MOFAPIError(t *testing.T) {
	repo := new(mockInvoiceRepo)
	mofClient := new(mockMOFClient)
	svc := newTestInvoiceService(repo, mofClient)

	mofClient.On("FetchInvoices", "/ABCD123", "2026/01/01", "2026/01/31").
		Return(nil, assert.AnError)

	_, err := svc.SyncInvoices(1, "/ABCD123", "2026/01/01", "2026/01/31")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MOF API")
}

func TestSyncInvoices_WithDuplicates(t *testing.T) {
	repo := new(mockInvoiceRepo)
	mofClient := new(mockMOFClient)
	svc := newTestInvoiceService(repo, mofClient)

	mofResp := &MOFResponse{
		Version: "0.5",
		Code:    200,
		Message: "成功",
		Details: []MOFInvoice{
			{
				InvNum:     "AB12345678",
				SellerName: "全家",
				Amount:     json.Number("100"),
				InvDate:    "2026/01/05 14:30:00",
			},
		},
	}

	mofClient.On("FetchInvoices", "/ABCD123", "2026/01/01", "2026/01/31").Return(mofResp, nil)
	// BatchCreate returns 0 (all duplicates)
	repo.On("BatchCreate", mock.Anything).Return(0, nil)

	created, err := svc.SyncInvoices(1, "/ABCD123", "2026/01/01", "2026/01/31")
	assert.NoError(t, err)
	assert.Equal(t, 0, created)
}

func TestListInvoices(t *testing.T) {
	repo := new(mockInvoiceRepo)
	svc := newTestInvoiceService(repo, nil)

	invoices := []models.Invoice{
		{ID: 1, InvoiceNumber: "AB12345678", Amount: 100},
		{ID: 2, InvoiceNumber: "CD87654321", Amount: 200},
	}
	repo.On("List", uint(1), (*time.Time)(nil), (*time.Time)(nil), 1, 20).Return(invoices, int64(2), nil)

	result, total, err := svc.ListInvoices(1, nil, nil, 1, 20)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), total)
}

func TestConfirmDuplicate(t *testing.T) {
	repo := new(mockInvoiceRepo)
	svc := newTestInvoiceService(repo, nil)

	invoice := &models.Invoice{ID: 1, InvoiceNumber: "AB12345678", Amount: 100}
	repo.On("GetByID", uint(1)).Return(invoice, nil)
	repo.On("Update", mock.AnythingOfType("*models.Invoice")).Return(nil)

	err := svc.ConfirmDuplicate(1, 42)
	assert.NoError(t, err)

	updatedInvoice := repo.Calls[1].Arguments.Get(0).(*models.Invoice)
	assert.True(t, updatedInvoice.IsDuplicated)
	assert.Equal(t, uint(42), *updatedInvoice.DuplicatedTransactionID)
	assert.Equal(t, 1.0, *updatedInvoice.ConfidenceScore)
}

func TestDeleteInvoice(t *testing.T) {
	repo := new(mockInvoiceRepo)
	svc := newTestInvoiceService(repo, nil)

	repo.On("Delete", uint(1)).Return(nil)

	err := svc.DeleteInvoice(1)
	assert.NoError(t, err)
	repo.AssertCalled(t, "Delete", uint(1))
}

func TestConvertMOFInvoice(t *testing.T) {
	svc := NewInvoiceService(nil, "", "")

	mof := MOFInvoice{
		InvNum:     "AB12345678",
		SellerName: "全家便利商店",
		Amount:     json.Number("150.50"),
		InvDate:    "2026/01/05 14:30:00",
		SellerBAN:  "12345678",
		InvStatus:  "已使用",
		Details:    json.RawMessage(`[{"description":"商品A","amount":"150.50"}]`),
	}

	invoice, err := svc.convertMOFInvoice(1, mof)
	assert.NoError(t, err)
	assert.Equal(t, "AB12345678", invoice.InvoiceNumber)
	assert.Equal(t, "全家便利商店", invoice.SellerName)
	assert.Equal(t, 150.50, invoice.Amount)
	assert.Equal(t, "12345678", invoice.SellerBAN)
	assert.Equal(t, uint(1), invoice.UserID)
	assert.Equal(t, 2026, invoice.InvoiceDate.Year())
	assert.Equal(t, time.January, invoice.InvoiceDate.Month())
	assert.Equal(t, 5, invoice.InvoiceDate.Day())
}

func TestConvertMOFInvoice_AlternativeDateFormat(t *testing.T) {
	svc := NewInvoiceService(nil, "", "")

	mof := MOFInvoice{
		InvNum:  "AB12345678",
		Amount:  json.Number("100"),
		InvDate: "2026/01/05",
	}

	invoice, err := svc.convertMOFInvoice(1, mof)
	assert.NoError(t, err)
	assert.Equal(t, 5, invoice.InvoiceDate.Day())
}

func TestConvertMOFInvoice_InvalidAmount(t *testing.T) {
	svc := NewInvoiceService(nil, "", "")

	mof := MOFInvoice{
		InvNum:  "AB12345678",
		Amount:  json.Number("invalid"),
		InvDate: "2026/01/05",
	}

	_, err := svc.convertMOFInvoice(1, mof)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid amount")
}
