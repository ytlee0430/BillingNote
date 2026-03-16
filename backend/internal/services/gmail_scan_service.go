package services

import (
	"billing-note/internal/models"
	"billing-note/internal/repository"
	"billing-note/pkg/errors"
	"billing-note/pkg/logger"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
	gmail "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// GmailAPIClient abstracts Gmail API calls for testability
type GmailAPIClient interface {
	ListMessages(query string, maxResults int64) ([]*gmail.Message, error)
	GetMessage(id string) (*gmail.Message, error)
	GetAttachment(messageID, attachmentID string) ([]byte, error)
}

// realGmailClient wraps the actual Google Gmail API
type realGmailClient struct {
	service *gmail.Service
}

func newRealGmailClient(ctx context.Context, oauthConfig *oauth2.Config, token *oauth2.Token) (*realGmailClient, error) {
	tokenSource := oauthConfig.TokenSource(ctx, token)
	svc, err := gmail.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, err
	}
	return &realGmailClient{service: svc}, nil
}

func (c *realGmailClient) ListMessages(query string, maxResults int64) ([]*gmail.Message, error) {
	var allMessages []*gmail.Message
	call := c.service.Users.Messages.List("me").Q(query).MaxResults(maxResults)

	resp, err := call.Do()
	if err != nil {
		return nil, err
	}
	allMessages = append(allMessages, resp.Messages...)

	// Handle pagination
	for resp.NextPageToken != "" {
		resp, err = call.PageToken(resp.NextPageToken).Do()
		if err != nil {
			return nil, err
		}
		allMessages = append(allMessages, resp.Messages...)
	}

	return allMessages, nil
}

func (c *realGmailClient) GetMessage(id string) (*gmail.Message, error) {
	return c.service.Users.Messages.Get("me", id).Do()
}

func (c *realGmailClient) GetAttachment(messageID, attachmentID string) ([]byte, error) {
	att, err := c.service.Users.Messages.Attachments.Get("me", messageID, attachmentID).Do()
	if err != nil {
		return nil, err
	}
	return base64.URLEncoding.DecodeString(att.Data)
}

// ScanResult represents the result of a Gmail scan
type ScanResult struct {
	Scanned      int            `json:"scanned"`
	Downloaded   int            `json:"downloaded"`
	AutoParsed   int            `json:"auto_parsed"`
	Imported     int            `json:"imported"`
	Failed       int            `json:"failed"`
	ParseResults []UploadResult `json:"parse_results,omitempty"`
	Status       string         `json:"status"`
	ErrorMessage string         `json:"error_message,omitempty"`
}

// GmailScanService handles Gmail email scanning and PDF download
type GmailScanService struct {
	gmailService  *GmailService
	uploadService *UploadService
	repo          repository.GmailRepository
	uploadDir     string
	// For testing: injectable client factory
	clientFactory func(ctx context.Context, token *oauth2.Token) (GmailAPIClient, error)
}

// NewGmailScanService creates a new Gmail scan service
func NewGmailScanService(
	gmailService *GmailService,
	uploadService *UploadService,
	repo repository.GmailRepository,
	uploadDir string,
) *GmailScanService {
	s := &GmailScanService{
		gmailService:  gmailService,
		uploadService: uploadService,
		repo:          repo,
		uploadDir:     uploadDir,
	}
	// Default client factory uses real Gmail API
	s.clientFactory = func(ctx context.Context, token *oauth2.Token) (GmailAPIClient, error) {
		return newRealGmailClient(ctx, gmailService.oauthConfig, token)
	}
	return s
}

// SetClientFactory allows injecting a mock client factory for testing
func (s *GmailScanService) SetClientFactory(factory func(ctx context.Context, token *oauth2.Token) (GmailAPIClient, error)) {
	s.clientFactory = factory
}

// TriggerScan executes a Gmail scan for the given user
func (s *GmailScanService) TriggerScan(userID uint) (*ScanResult, error) {
	log := logger.ServiceLog("GmailScanService", "TriggerScan")

	// Get OAuth token
	oauthToken, err := s.gmailService.GetOAuthTokenForUser(userID)
	if err != nil {
		log.WithFields(logger.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to get OAuth token for scan")
		return nil, errors.NewInternalError("Gmail not connected", err)
	}

	// Get scan rules
	rule, err := s.gmailService.GetSettings(userID)
	if err != nil {
		return nil, errors.NewInternalError("Failed to get scan settings", err)
	}

	// Build Gmail query
	query := s.buildQuery(rule)

	log.WithFields(logger.Fields{
		"user_id": userID,
		"query":   query,
	}).Info("Starting Gmail scan")

	// Create Gmail API client
	ctx := context.Background()
	client, err := s.clientFactory(ctx, oauthToken)
	if err != nil {
		log.WithError(err).Error("Failed to create Gmail API client")
		s.recordScanHistory(userID, 0, 0, "error", "Failed to create Gmail client: "+err.Error())
		return nil, errors.NewInternalError("Failed to create Gmail client", err)
	}

	// Search for matching emails
	messages, err := client.ListMessages(query, 50)
	if err != nil {
		log.WithError(err).Error("Failed to search Gmail")
		s.recordScanHistory(userID, 0, 0, "error", "Gmail search failed: "+err.Error())
		return nil, errors.NewInternalError("Failed to search Gmail", err)
	}

	scanned := len(messages)
	downloaded := 0
	autoParsed := 0
	totalImported := 0
	failed := 0
	var parseResults []UploadResult

	log.WithFields(logger.Fields{
		"user_id":      userID,
		"emails_found": scanned,
	}).Info("Gmail search completed")

	// Process each email
	for _, msg := range messages {
		fullMsg, err := client.GetMessage(msg.Id)
		if err != nil {
			log.WithFields(logger.Fields{
				"user_id":    userID,
				"message_id": msg.Id,
				"error":      err.Error(),
			}).Warn("Failed to get email message, skipping")
			continue
		}

		// Extract PDF attachments
		pdfPaths, err := s.downloadPDFAttachments(client, userID, fullMsg)
		if err != nil {
			log.WithFields(logger.Fields{
				"user_id":    userID,
				"message_id": msg.Id,
				"error":      err.Error(),
			}).Warn("Failed to download attachments, skipping")
			continue
		}

		downloaded += len(pdfPaths)

		// Parse downloaded PDFs using existing pipeline
		if s.uploadService != nil {
			for _, pdfPath := range pdfPaths {
				result, err := s.uploadService.ParsePDF(userID, pdfPath)
				if err != nil {
					log.WithFields(logger.Fields{
						"user_id":  userID,
						"pdf_path": pdfPath,
						"error":    err.Error(),
					}).Warn("Failed to parse downloaded PDF")
					parseResults = append(parseResults, UploadResult{
						Filename: filepath.Base(pdfPath),
						Error:    err.Error(),
					})
					failed++
					continue
				}
				if result.Error != "" {
					failed++
				} else {
					// Auto-import parsed transactions into database
					imported, importErr := s.uploadService.ImportTransactions(userID, result.Transactions)
					if importErr != nil {
						log.WithFields(logger.Fields{
							"user_id":  userID,
							"pdf_path": pdfPath,
							"error":    importErr.Error(),
						}).Warn("Failed to import transactions from parsed PDF")
					} else {
						totalImported += imported
						if imported > 0 {
							log.WithFields(logger.Fields{
								"user_id":  userID,
								"pdf_path": pdfPath,
								"imported": imported,
							}).Info("Auto-imported transactions from Gmail PDF")
						}
					}
					autoParsed++
				}
				parseResults = append(parseResults, *result)
			}
		}
	}

	// Update last scan time
	rule.LastScanAt = timePtr(time.Now())
	_ = s.repo.SaveScanRule(rule)

	// Record scan history
	status := "completed"
	errMsg := ""
	if downloaded == 0 && scanned > 0 {
		status = "no_pdfs"
		errMsg = "Emails found but no PDF attachments"
	}
	s.recordScanHistory(userID, scanned, downloaded, status, errMsg)

	log.WithFields(logger.Fields{
		"user_id":            userID,
		"emails_found":       scanned,
		"pdfs_downloaded":    downloaded,
		"auto_parsed":        autoParsed,
		"total_imported":     totalImported,
		"failed":             failed,
	}).Info("Gmail scan completed")

	return &ScanResult{
		Scanned:      scanned,
		Downloaded:   downloaded,
		AutoParsed:   autoParsed,
		Imported:     totalImported,
		Failed:       failed,
		ParseResults: parseResults,
		Status:       status,
		ErrorMessage: errMsg,
	}, nil
}

// GetScanHistory returns recent scan history for a user
func (s *GmailScanService) GetScanHistory(userID uint, limit int) ([]models.GmailScanHistory, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.ListScanHistory(userID, limit)
}

// --- Internal helpers ---

func (s *GmailScanService) buildQuery(rule *models.GmailScanRule) string {
	var parts []string

	// Combine sender and subject keywords into one OR group
	// Gmail syntax: {from:X from:Y subject:A subject:B} matches any of these
	var orParts []string
	for _, kw := range rule.SenderKeywords {
		orParts = append(orParts, fmt.Sprintf("from:%s", kw))
	}
	for _, kw := range rule.SubjectKeywords {
		orParts = append(orParts, fmt.Sprintf("subject:%s", kw))
	}
	if len(orParts) > 0 {
		parts = append(parts, "{"+strings.Join(orParts, " ")+"}")
	}

	// Require attachment
	if rule.RequireAttachment {
		parts = append(parts, "has:attachment")
	}

	// Time range: use last scan time, or default to 6 months back
	if rule.LastScanAt != nil {
		parts = append(parts, fmt.Sprintf("after:%s", rule.LastScanAt.Format("2006/01/02")))
	} else {
		sixMonthsAgo := time.Now().AddDate(0, -6, 0)
		parts = append(parts, fmt.Sprintf("after:%s", sixMonthsAgo.Format("2006/01/02")))
	}

	return strings.Join(parts, " ")
}

func (s *GmailScanService) downloadPDFAttachments(client GmailAPIClient, userID uint, msg *gmail.Message) ([]string, error) {
	var pdfPaths []string

	parts := getAllParts(msg.Payload)
	for _, part := range parts {
		// Only process PDF attachments
		filename := part.Filename
		if filename == "" || !strings.HasSuffix(strings.ToLower(filename), ".pdf") {
			continue
		}

		if part.Body == nil || part.Body.AttachmentId == "" {
			continue
		}

		// Download attachment
		data, err := client.GetAttachment(msg.Id, part.Body.AttachmentId)
		if err != nil {
			return pdfPaths, fmt.Errorf("failed to download attachment %s: %w", filename, err)
		}

		// Save to user's gmail upload directory
		dir := filepath.Join(s.uploadDir, fmt.Sprintf("%d", userID), "gmail")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return pdfPaths, fmt.Errorf("failed to create directory: %w", err)
		}

		// Use message ID + filename to prevent duplicates
		safeFilename := fmt.Sprintf("%s_%s", msg.Id, sanitizeFilename(filename))
		filePath := filepath.Join(dir, safeFilename)

		// Skip download if already exists, but still include for parsing
		if _, err := os.Stat(filePath); err == nil {
			pdfPaths = append(pdfPaths, filePath)
			continue
		}

		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return pdfPaths, fmt.Errorf("failed to save attachment: %w", err)
		}

		pdfPaths = append(pdfPaths, filePath)
	}

	return pdfPaths, nil
}

func getAllParts(part *gmail.MessagePart) []*gmail.MessagePart {
	var parts []*gmail.MessagePart
	if part == nil {
		return parts
	}
	parts = append(parts, part)
	for _, p := range part.Parts {
		parts = append(parts, getAllParts(p)...)
	}
	return parts
}

func sanitizeFilename(name string) string {
	// Replace problematic characters
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", " ", "_")
	return replacer.Replace(name)
}

func (s *GmailScanService) recordScanHistory(userID uint, emailsFound, pdfsDownloaded int, status, errMsg string) {
	history := &models.GmailScanHistory{
		UserID:         userID,
		ScanAt:         time.Now(),
		EmailsFound:    emailsFound,
		PDFsDownloaded: pdfsDownloaded,
		Status:         status,
		ErrorMessage:   errMsg,
	}
	if err := s.repo.CreateScanHistory(history); err != nil {
		logger.WithError(err).Warn("Failed to record scan history")
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
