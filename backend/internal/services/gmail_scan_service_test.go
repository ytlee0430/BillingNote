package services

import (
	"billing-note/internal/models"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
	gmail "google.golang.org/api/gmail/v1"
	"gorm.io/gorm"
)

// --- Mock Gmail API Client ---

type mockGmailAPIClient struct {
	mock.Mock
}

func (m *mockGmailAPIClient) ListMessages(query string, maxResults int64) ([]*gmail.Message, error) {
	args := m.Called(query, maxResults)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*gmail.Message), args.Error(1)
}

func (m *mockGmailAPIClient) GetMessage(id string) (*gmail.Message, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmail.Message), args.Error(1)
}

func (m *mockGmailAPIClient) GetAttachment(messageID, attachmentID string) ([]byte, error) {
	args := m.Called(messageID, attachmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// --- Tests ---

func newTestScanService(t *testing.T, repo *mockGmailRepo, client *mockGmailAPIClient) (*GmailScanService, *GmailService) {
	gmailSvc := newTestGmailService(repo)

	uploadDir := t.TempDir()
	scanSvc := NewGmailScanService(gmailSvc, nil, repo, uploadDir)
	if client != nil {
		scanSvc.SetClientFactory(func(ctx context.Context, token *oauth2.Token) (GmailAPIClient, error) {
			return client, nil
		})
	}

	return scanSvc, gmailSvc
}

func TestBuildQuery_Default(t *testing.T) {
	repo := new(mockGmailRepo)
	scanSvc, _ := newTestScanService(t, repo, nil)

	rule := &models.GmailScanRule{
		SenderKeywords:    []string{"credit", "信用卡"},
		SubjectKeywords:   []string{"帳單", "statement"},
		RequireAttachment: true,
	}

	query := scanSvc.buildQuery(rule)
	assert.Contains(t, query, "from:credit")
	assert.Contains(t, query, "from:信用卡")
	assert.Contains(t, query, "subject:帳單")
	assert.Contains(t, query, "subject:statement")
	assert.Contains(t, query, "has:attachment")
}

func TestBuildQuery_WithLastScan(t *testing.T) {
	repo := new(mockGmailRepo)
	scanSvc, _ := newTestScanService(t, repo, nil)

	lastScan := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	rule := &models.GmailScanRule{
		SenderKeywords:    []string{"credit"},
		SubjectKeywords:   []string{"帳單"},
		RequireAttachment: true,
		LastScanAt:        &lastScan,
	}

	query := scanSvc.buildQuery(rule)
	assert.Contains(t, query, "after:2026/03/01")
}

func TestBuildQuery_NoAttachmentRequired(t *testing.T) {
	repo := new(mockGmailRepo)
	scanSvc, _ := newTestScanService(t, repo, nil)

	rule := &models.GmailScanRule{
		SenderKeywords:    []string{"credit"},
		RequireAttachment: false,
	}

	query := scanSvc.buildQuery(rule)
	assert.NotContains(t, query, "has:attachment")
}

func TestTriggerScan_NotConnected(t *testing.T) {
	repo := new(mockGmailRepo)
	repo.On("GetToken", uint(1)).Return(nil, gorm.ErrRecordNotFound)

	scanSvc, _ := newTestScanService(t, repo, nil)

	result, err := scanSvc.TriggerScan(1)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestTriggerScan_NoEmails(t *testing.T) {
	repo := new(mockGmailRepo)
	client := new(mockGmailAPIClient)
	scanSvc, gmailSvc := newTestScanService(t, repo, client)

	// Setup: user has valid tokens
	accessEnc, _ := gmailSvc.crypto.Encrypt("test-access-token")
	refreshEnc, _ := gmailSvc.crypto.Encrypt("test-refresh-token")
	futureExpiry := time.Now().Add(1 * time.Hour)

	token := &models.GmailToken{
		UserID:                1,
		AccessTokenEncrypted:  accessEnc,
		RefreshTokenEncrypted: refreshEnc,
		TokenExpiry:           &futureExpiry,
	}
	repo.On("GetToken", uint(1)).Return(token, nil)
	repo.On("GetScanRule", uint(1)).Return(nil, gorm.ErrRecordNotFound)
	repo.On("SaveScanRule", mock.Anything).Return(nil)
	repo.On("CreateScanHistory", mock.Anything).Return(nil)

	// No emails found
	client.On("ListMessages", mock.Anything, int64(50)).Return([]*gmail.Message{}, nil)

	result, err := scanSvc.TriggerScan(1)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.EmailsFound)
	assert.Equal(t, 0, result.PDFsDownloaded)
	assert.Equal(t, "completed", result.Status)
}

func TestTriggerScan_WithPDFAttachment(t *testing.T) {
	repo := new(mockGmailRepo)
	client := new(mockGmailAPIClient)
	scanSvc, gmailSvc := newTestScanService(t, repo, client)

	accessEnc, _ := gmailSvc.crypto.Encrypt("test-access-token")
	refreshEnc, _ := gmailSvc.crypto.Encrypt("test-refresh-token")
	futureExpiry := time.Now().Add(1 * time.Hour)

	token := &models.GmailToken{
		UserID:                1,
		AccessTokenEncrypted:  accessEnc,
		RefreshTokenEncrypted: refreshEnc,
		TokenExpiry:           &futureExpiry,
	}
	repo.On("GetToken", uint(1)).Return(token, nil)
	repo.On("GetScanRule", uint(1)).Return(nil, gorm.ErrRecordNotFound)
	repo.On("SaveScanRule", mock.Anything).Return(nil)
	repo.On("CreateScanHistory", mock.Anything).Return(nil)

	// One email with PDF attachment
	messages := []*gmail.Message{{Id: "msg1"}}
	client.On("ListMessages", mock.Anything, int64(50)).Return(messages, nil)

	fullMsg := &gmail.Message{
		Id: "msg1",
		Payload: &gmail.MessagePart{
			Parts: []*gmail.MessagePart{
				{
					Filename: "statement.pdf",
					MimeType: "application/pdf",
					Body: &gmail.MessagePartBody{
						AttachmentId: "att1",
						Size:         1024,
					},
				},
			},
		},
	}
	client.On("GetMessage", "msg1").Return(fullMsg, nil)
	client.On("GetAttachment", "msg1", "att1").Return([]byte("%PDF-1.4 fake pdf content"), nil)

	result, err := scanSvc.TriggerScan(1)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.EmailsFound)
	assert.Equal(t, 1, result.PDFsDownloaded)

	client.AssertCalled(t, "GetAttachment", "msg1", "att1")
}

func TestTriggerScan_SkipNonPDFAttachments(t *testing.T) {
	repo := new(mockGmailRepo)
	client := new(mockGmailAPIClient)
	scanSvc, gmailSvc := newTestScanService(t, repo, client)

	accessEnc, _ := gmailSvc.crypto.Encrypt("test-access-token")
	refreshEnc, _ := gmailSvc.crypto.Encrypt("test-refresh-token")
	futureExpiry := time.Now().Add(1 * time.Hour)

	token := &models.GmailToken{
		UserID:                1,
		AccessTokenEncrypted:  accessEnc,
		RefreshTokenEncrypted: refreshEnc,
		TokenExpiry:           &futureExpiry,
	}
	repo.On("GetToken", uint(1)).Return(token, nil)
	repo.On("GetScanRule", uint(1)).Return(nil, gorm.ErrRecordNotFound)
	repo.On("SaveScanRule", mock.Anything).Return(nil)
	repo.On("CreateScanHistory", mock.Anything).Return(nil)

	messages := []*gmail.Message{{Id: "msg1"}}
	client.On("ListMessages", mock.Anything, int64(50)).Return(messages, nil)

	// Email with non-PDF attachment only
	fullMsg := &gmail.Message{
		Id: "msg1",
		Payload: &gmail.MessagePart{
			Parts: []*gmail.MessagePart{
				{
					Filename: "image.jpg",
					MimeType: "image/jpeg",
					Body: &gmail.MessagePartBody{
						AttachmentId: "att1",
						Size:         2048,
					},
				},
			},
		},
	}
	client.On("GetMessage", "msg1").Return(fullMsg, nil)

	result, err := scanSvc.TriggerScan(1)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.EmailsFound)
	assert.Equal(t, 0, result.PDFsDownloaded)

	// Should NOT have called GetAttachment since it's not a PDF
	client.AssertNotCalled(t, "GetAttachment", mock.Anything, mock.Anything)
}

func TestGetScanHistory(t *testing.T) {
	repo := new(mockGmailRepo)
	scanSvc, _ := newTestScanService(t, repo, nil)

	history := []models.GmailScanHistory{
		{ID: 1, UserID: 1, EmailsFound: 5, PDFsDownloaded: 2, Status: "completed"},
		{ID: 2, UserID: 1, EmailsFound: 0, PDFsDownloaded: 0, Status: "completed"},
	}
	repo.On("ListScanHistory", uint(1), 20).Return(history, nil)

	result, err := scanSvc.GetScanHistory(1, 0)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 5, result[0].EmailsFound)
}

func TestSanitizeFilename(t *testing.T) {
	assert.Equal(t, "file_name.pdf", sanitizeFilename("file name.pdf"))
	assert.Equal(t, "path_to_file.pdf", sanitizeFilename("path/to/file.pdf"))
	assert.Equal(t, "file_test.pdf", sanitizeFilename("file:test.pdf"))
}

func TestGetAllParts_Nested(t *testing.T) {
	part := &gmail.MessagePart{
		MimeType: "multipart/mixed",
		Parts: []*gmail.MessagePart{
			{
				MimeType: "text/plain",
			},
			{
				MimeType: "multipart/alternative",
				Parts: []*gmail.MessagePart{
					{
						Filename: "statement.pdf",
						MimeType: "application/pdf",
						Body: &gmail.MessagePartBody{
							AttachmentId: "att1",
						},
					},
				},
			},
		},
	}

	allParts := getAllParts(part)
	assert.Len(t, allParts, 4) // root + 2 children + 1 grandchild

	// Find the PDF part
	pdfFound := false
	for _, p := range allParts {
		if p.Filename == "statement.pdf" {
			pdfFound = true
			break
		}
	}
	assert.True(t, pdfFound)
}
