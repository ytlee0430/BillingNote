package repository

import (
	"billing-note/internal/models"

	"gorm.io/gorm"
)

// GmailRepository defines the interface for Gmail data access
type GmailRepository interface {
	// Token operations
	SaveToken(token *models.GmailToken) error
	GetToken(userID uint) (*models.GmailToken, error)
	DeleteToken(userID uint) error

	// Scan rule operations
	GetScanRule(userID uint) (*models.GmailScanRule, error)
	SaveScanRule(rule *models.GmailScanRule) error

	// Scan history operations
	CreateScanHistory(history *models.GmailScanHistory) error
	ListScanHistory(userID uint, limit int) ([]models.GmailScanHistory, error)
}

type gmailRepository struct {
	db *gorm.DB
}

func NewGmailRepository(db *gorm.DB) GmailRepository {
	return &gmailRepository{db: db}
}

func (r *gmailRepository) SaveToken(token *models.GmailToken) error {
	var existing models.GmailToken
	err := r.db.Where("user_id = ?", token.UserID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(token).Error
	}
	if err != nil {
		return err
	}
	existing.AccessTokenEncrypted = token.AccessTokenEncrypted
	existing.RefreshTokenEncrypted = token.RefreshTokenEncrypted
	existing.TokenExpiry = token.TokenExpiry
	existing.Scopes = token.Scopes
	return r.db.Save(&existing).Error
}

func (r *gmailRepository) GetToken(userID uint) (*models.GmailToken, error) {
	var token models.GmailToken
	err := r.db.Where("user_id = ?", userID).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *gmailRepository) DeleteToken(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.GmailToken{}).Error
}

func (r *gmailRepository) GetScanRule(userID uint) (*models.GmailScanRule, error) {
	var rule models.GmailScanRule
	err := r.db.Where("user_id = ?", userID).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *gmailRepository) SaveScanRule(rule *models.GmailScanRule) error {
	var existing models.GmailScanRule
	err := r.db.Where("user_id = ?", rule.UserID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(rule).Error
	}
	if err != nil {
		return err
	}
	existing.Enabled = rule.Enabled
	existing.SenderKeywords = rule.SenderKeywords
	existing.SubjectKeywords = rule.SubjectKeywords
	existing.RequireAttachment = rule.RequireAttachment
	existing.LastScanAt = rule.LastScanAt
	return r.db.Save(&existing).Error
}

func (r *gmailRepository) CreateScanHistory(history *models.GmailScanHistory) error {
	return r.db.Create(history).Error
}

func (r *gmailRepository) ListScanHistory(userID uint, limit int) ([]models.GmailScanHistory, error) {
	var history []models.GmailScanHistory
	query := r.db.Where("user_id = ?", userID).Order("scan_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&history).Error
	return history, err
}
