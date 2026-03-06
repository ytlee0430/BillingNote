package models

import (
	"time"

	"github.com/lib/pq"
)

// GmailToken stores encrypted OAuth tokens for Gmail integration
type GmailToken struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	UserID                uint      `gorm:"not null;uniqueIndex" json:"user_id"`
	AccessTokenEncrypted  string    `gorm:"not null" json:"-"`
	RefreshTokenEncrypted string    `gorm:"not null" json:"-"`
	TokenExpiry           *time.Time `json:"token_expiry,omitempty"`
	Scopes                string    `json:"scopes,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (GmailToken) TableName() string {
	return "gmail_tokens"
}

// GmailScanRule stores per-user scan configuration
type GmailScanRule struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	UserID            uint           `gorm:"not null;uniqueIndex" json:"user_id"`
	Enabled           bool           `gorm:"default:false" json:"enabled"`
	SenderKeywords    pq.StringArray `gorm:"type:text[];default:'{\"credit\",\"信用卡\",\"帳單\",\"statement\"}'" json:"sender_keywords"`
	SubjectKeywords   pq.StringArray `gorm:"type:text[];default:'{\"帳單\",\"電子帳單\",\"statement\"}'" json:"subject_keywords"`
	RequireAttachment bool           `gorm:"default:true" json:"require_attachment"`
	LastScanAt        *time.Time     `json:"last_scan_at,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (GmailScanRule) TableName() string {
	return "gmail_scan_rules"
}

// GmailScanHistory records each scan attempt
type GmailScanHistory struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"not null;index" json:"user_id"`
	ScanAt         time.Time `gorm:"default:now()" json:"scan_at"`
	EmailsFound    int       `gorm:"default:0" json:"emails_found"`
	PDFsDownloaded int       `gorm:"default:0" json:"pdfs_downloaded"`
	Status         string    `gorm:"default:'completed';size:20" json:"status"`
	ErrorMessage   string    `json:"error_message,omitempty"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (GmailScanHistory) TableName() string {
	return "gmail_scan_history"
}

// GmailStatusResponse represents the Gmail connection status
type GmailStatusResponse struct {
	Connected  bool       `json:"connected"`
	Email      string     `json:"email,omitempty"`
	Scopes     string     `json:"scopes,omitempty"`
	LastScanAt *time.Time `json:"last_scan_at,omitempty"`
	ConnectedAt *time.Time `json:"connected_at,omitempty"`
}

// GmailSettingsInput represents input for updating Gmail scan settings
type GmailSettingsInput struct {
	Enabled           *bool    `json:"enabled"`
	SenderKeywords    []string `json:"sender_keywords"`
	SubjectKeywords   []string `json:"subject_keywords"`
	RequireAttachment *bool    `json:"require_attachment"`
}
