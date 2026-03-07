package services

import (
	"billing-note/internal/models"
	"billing-note/internal/repository"
	"billing-note/pkg/errors"
	"billing-note/pkg/logger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// MOFAPIClient abstracts the MOF e-invoice API for testability
type MOFAPIClient interface {
	FetchInvoices(carrierCode, startDate, endDate string) (*MOFResponse, error)
}

// MOFResponse represents the response from the MOF API
type MOFResponse struct {
	Version string       `json:"v"`
	Code    int          `json:"code"`
	Message string       `json:"msg"`
	Details []MOFInvoice `json:"details"`
}

// MOFInvoice represents a single invoice from the MOF API
type MOFInvoice struct {
	InvNum      string          `json:"invNum"`
	CardType    string          `json:"cardType"`
	CardNo      string          `json:"cardNo"`
	SellerName  string          `json:"sellerName"`
	InvStatus   string          `json:"invStatus"`
	Amount      json.Number     `json:"amount"`
	InvPeriod   string          `json:"invPeriod"`
	InvDate     string          `json:"invDate"`
	SellerBAN   string          `json:"sellerBan"`
	InvoiceTime string          `json:"invoiceTime"`
	Details     json.RawMessage `json:"details"`
}

// realMOFClient implements MOFAPIClient with actual HTTP calls
type realMOFClient struct {
	apiURL string
	appID  string
}

func newRealMOFClient(apiURL, appID string) *realMOFClient {
	return &realMOFClient{apiURL: apiURL, appID: appID}
}

func (c *realMOFClient) FetchInvoices(carrierCode, startDate, endDate string) (*MOFResponse, error) {
	now := time.Now()
	timestamp := fmt.Sprintf("%d", now.Unix())

	params := url.Values{}
	params.Set("version", "0.5")
	params.Set("action", "carrierInvChk")
	params.Set("cardType", "3J0002")
	params.Set("cardNo", carrierCode)
	params.Set("expTimeStamp", timestamp)
	params.Set("timeStamp", timestamp)
	params.Set("startDate", startDate)
	params.Set("endDate", endDate)
	params.Set("onlyWinningInv", "N")
	params.Set("uuid", c.appID)
	params.Set("appID", c.appID)

	reqURL := fmt.Sprintf("%s?%s", c.apiURL, params.Encode())
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("MOF API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read MOF API response: %w", err)
	}

	var mofResp MOFResponse
	if err := json.Unmarshal(body, &mofResp); err != nil {
		return nil, fmt.Errorf("failed to parse MOF API response: %w", err)
	}

	if mofResp.Code != 200 {
		return nil, fmt.Errorf("MOF API error: %s (code: %d)", mofResp.Message, mofResp.Code)
	}

	return &mofResp, nil
}

// InvoiceService handles invoice operations
type InvoiceService struct {
	repo      repository.InvoiceRepository
	mofClient MOFAPIClient
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(
	repo repository.InvoiceRepository,
	apiURL, appID string,
) *InvoiceService {
	svc := &InvoiceService{
		repo:      repo,
		mofClient: newRealMOFClient(apiURL, appID),
	}
	return svc
}

// SetMOFClient allows injecting a mock MOF API client for testing
func (s *InvoiceService) SetMOFClient(client MOFAPIClient) {
	s.mofClient = client
}

// SyncInvoices syncs invoices from the MOF API
func (s *InvoiceService) SyncInvoices(userID uint, carrierCode, startDate, endDate string) (int, error) {
	log := logger.ServiceLog("InvoiceService", "SyncInvoices")

	if carrierCode == "" {
		return 0, errors.NewValidationError("Invoice carrier code is required")
	}

	// Validate carrier code format
	if !strings.HasPrefix(carrierCode, "/") || len(carrierCode) != 8 {
		return 0, errors.NewValidationError("Invalid carrier code format. Must be /XXXXXXX (7 characters after /)")
	}

	log.WithFields(logger.Fields{
		"user_id":    userID,
		"start_date": startDate,
		"end_date":   endDate,
	}).Info("Starting invoice sync from MOF API")

	// Fetch from MOF API
	resp, err := s.mofClient.FetchInvoices(carrierCode, startDate, endDate)
	if err != nil {
		log.WithError(err).Error("Failed to fetch invoices from MOF API")
		return 0, errors.NewInternalError("Failed to fetch invoices from MOF API", err)
	}

	// Convert to model and store
	invoices := make([]*models.Invoice, 0, len(resp.Details))
	for _, detail := range resp.Details {
		invoice, err := s.convertMOFInvoice(userID, detail)
		if err != nil {
			log.WithFields(logger.Fields{
				"invoice_number": detail.InvNum,
				"error":          err.Error(),
			}).Warn("Failed to convert invoice, skipping")
			continue
		}
		invoices = append(invoices, invoice)
	}

	// Batch create (skips duplicates)
	created, err := s.repo.BatchCreate(invoices)
	if err != nil {
		log.WithError(err).Error("Failed to store invoices")
		return 0, errors.NewDBError("store invoices", err)
	}

	log.WithFields(logger.Fields{
		"user_id":    userID,
		"fetched":    len(resp.Details),
		"created":    created,
		"duplicates": len(resp.Details) - created,
	}).Info("Invoice sync completed")

	return created, nil
}

// ListInvoices returns invoices with pagination
func (s *InvoiceService) ListInvoices(userID uint, startDate, endDate *time.Time, page, pageSize int) ([]models.Invoice, int64, error) {
	return s.repo.List(userID, startDate, endDate, page, pageSize)
}

// GetInvoice returns a single invoice by ID
func (s *InvoiceService) GetInvoice(id uint) (*models.Invoice, error) {
	return s.repo.GetByID(id)
}

// ConfirmDuplicate marks an invoice as a confirmed duplicate of a transaction
func (s *InvoiceService) ConfirmDuplicate(invoiceID, transactionID uint) error {
	log := logger.ServiceLog("InvoiceService", "ConfirmDuplicate")

	invoice, err := s.repo.GetByID(invoiceID)
	if err != nil {
		return errors.NewNotFoundError("Invoice", invoiceID)
	}

	invoice.IsDuplicated = true
	invoice.DuplicatedTransactionID = &transactionID
	confidence := 1.0
	invoice.ConfidenceScore = &confidence

	if err := s.repo.Update(invoice); err != nil {
		log.WithError(err).Error("Failed to confirm duplicate")
		return errors.NewDBError("confirm duplicate", err)
	}

	log.WithFields(logger.Fields{
		"invoice_id":     invoiceID,
		"transaction_id": transactionID,
	}).Info("Duplicate confirmed")

	return nil
}

// DeleteInvoice deletes an invoice
func (s *InvoiceService) DeleteInvoice(id uint) error {
	return s.repo.Delete(id)
}

// --- Internal helpers ---

func (s *InvoiceService) convertMOFInvoice(userID uint, mof MOFInvoice) (*models.Invoice, error) {
	// Parse amount
	amount, err := strconv.ParseFloat(string(mof.Amount), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	// Parse date - format: "2026/01/05 14:30:00"
	invoiceDate, err := time.Parse("2006/01/02 15:04:05", mof.InvDate)
	if err != nil {
		// Try alternative format
		invoiceDate, err = time.Parse("2006/01/02", mof.InvDate)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
	}

	return &models.Invoice{
		UserID:        userID,
		InvoiceNumber: mof.InvNum,
		InvoiceDate:   invoiceDate,
		SellerName:    mof.SellerName,
		SellerBAN:     mof.SellerBAN,
		Amount:        amount,
		Status:        mof.InvStatus,
		Items:         mof.Details,
	}, nil
}
