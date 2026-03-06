package services

import (
	"billing-note/internal/models"
	"billing-note/internal/repository"
	"billing-note/pkg/crypto"
	"billing-note/pkg/errors"
	"billing-note/pkg/logger"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// GmailService handles Gmail OAuth and integration operations
type GmailService struct {
	repo        repository.GmailRepository
	crypto      *crypto.AESCrypto
	oauthConfig *oauth2.Config
	stateSecret string
}

// NewGmailService creates a new Gmail service
func NewGmailService(
	repo repository.GmailRepository,
	encryptionKey string,
	clientID, clientSecret, redirectURI string,
	stateSecret string,
) (*GmailService, error) {
	log := logger.ServiceLog("GmailService", "NewGmailService")

	aesCrypto, err := crypto.NewAESCrypto(encryptionKey)
	if err != nil {
		log.WithError(err).Error("Failed to initialize AES encryption for Gmail tokens")
		return nil, errors.NewEncryptionError("gmail initialization", err)
	}

	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes: []string{
			gmail.GmailReadonlyScope,
			gmail.GmailMetadataScope,
		},
		Endpoint: google.Endpoint,
	}

	log.Info("Gmail service initialized successfully")

	return &GmailService{
		repo:        repo,
		crypto:      aesCrypto,
		oauthConfig: oauthConfig,
		stateSecret: stateSecret,
	}, nil
}

// GetAuthURL generates the Google OAuth authorization URL
func (s *GmailService) GetAuthURL(userID uint) (string, error) {
	log := logger.ServiceLog("GmailService", "GetAuthURL")

	state := s.generateState(userID)

	url := s.oauthConfig.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)

	log.WithField("user_id", userID).Info("Generated Gmail OAuth authorization URL")
	return url, nil
}

// HandleCallback exchanges the authorization code for tokens and stores them
func (s *GmailService) HandleCallback(userID uint, code, state string) error {
	log := logger.ServiceLog("GmailService", "HandleCallback")

	// Verify state to prevent CSRF
	if !s.verifyState(state, userID) {
		log.WithField("user_id", userID).Warn("Invalid OAuth state parameter")
		return errors.NewValidationError("Invalid OAuth state parameter")
	}

	// Exchange code for tokens
	token, err := s.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.WithFields(logger.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to exchange OAuth code for tokens")
		return errors.NewInternalError("Failed to exchange OAuth code", err)
	}

	// Encrypt tokens
	accessTokenEnc, err := s.crypto.Encrypt(token.AccessToken)
	if err != nil {
		log.WithError(err).Error("Failed to encrypt access token")
		return errors.NewEncryptionError("access token encryption", err)
	}

	refreshTokenEnc, err := s.crypto.Encrypt(token.RefreshToken)
	if err != nil {
		log.WithError(err).Error("Failed to encrypt refresh token")
		return errors.NewEncryptionError("refresh token encryption", err)
	}

	// Store tokens
	gmailToken := &models.GmailToken{
		UserID:                userID,
		AccessTokenEncrypted:  accessTokenEnc,
		RefreshTokenEncrypted: refreshTokenEnc,
		TokenExpiry:           &token.Expiry,
		Scopes:                gmail.GmailReadonlyScope + " " + gmail.GmailMetadataScope,
	}

	if err := s.repo.SaveToken(gmailToken); err != nil {
		log.WithFields(logger.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to save Gmail tokens")
		return errors.NewDBError("save Gmail tokens", err)
	}

	// Create default scan rules if not exist
	_, err = s.repo.GetScanRule(userID)
	if err != nil {
		defaultRule := &models.GmailScanRule{
			UserID:            userID,
			Enabled:           false,
			SenderKeywords:    []string{"credit", "信用卡", "帳單", "statement"},
			SubjectKeywords:   []string{"帳單", "電子帳單", "statement"},
			RequireAttachment: true,
		}
		if saveErr := s.repo.SaveScanRule(defaultRule); saveErr != nil {
			log.WithError(saveErr).Warn("Failed to create default scan rules")
		}
	}

	log.WithField("user_id", userID).Info("Gmail OAuth tokens stored successfully")
	return nil
}

// GetStatus returns the current Gmail connection status
func (s *GmailService) GetStatus(userID uint) (*models.GmailStatusResponse, error) {
	log := logger.ServiceLog("GmailService", "GetStatus")

	token, err := s.repo.GetToken(userID)
	if err != nil {
		log.WithField("user_id", userID).Debug("No Gmail token found")
		return &models.GmailStatusResponse{Connected: false}, nil
	}

	status := &models.GmailStatusResponse{
		Connected:   true,
		Scopes:      token.Scopes,
		ConnectedAt: &token.CreatedAt,
	}

	// Get email from Gmail API
	oauthToken, err := s.getOAuthToken(token)
	if err == nil {
		email, emailErr := s.fetchGmailEmail(oauthToken)
		if emailErr == nil {
			status.Email = email
		}
	}

	// Get last scan time from scan rules
	rule, err := s.repo.GetScanRule(userID)
	if err == nil {
		status.LastScanAt = rule.LastScanAt
	}

	log.WithField("user_id", userID).Debug("Gmail status retrieved")
	return status, nil
}

// Disconnect removes Gmail tokens and revokes access
func (s *GmailService) Disconnect(userID uint) error {
	log := logger.ServiceLog("GmailService", "Disconnect")

	token, err := s.repo.GetToken(userID)
	if err != nil {
		log.WithField("user_id", userID).Warn("No Gmail token to disconnect")
		return errors.NewNotFoundError("Gmail connection", userID)
	}

	// Try to revoke the token
	oauthToken, err := s.getOAuthToken(token)
	if err == nil {
		s.revokeToken(oauthToken.AccessToken)
	}

	// Delete token from database
	if err := s.repo.DeleteToken(userID); err != nil {
		log.WithFields(logger.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to delete Gmail tokens")
		return errors.NewDBError("delete Gmail tokens", err)
	}

	log.WithField("user_id", userID).Info("Gmail disconnected successfully")
	return nil
}

// UpdateSettings updates Gmail scan settings
func (s *GmailService) UpdateSettings(userID uint, input models.GmailSettingsInput) error {
	log := logger.ServiceLog("GmailService", "UpdateSettings")

	rule, err := s.repo.GetScanRule(userID)
	if err != nil {
		rule = &models.GmailScanRule{UserID: userID}
	}

	if input.Enabled != nil {
		rule.Enabled = *input.Enabled
	}
	if input.SenderKeywords != nil {
		rule.SenderKeywords = input.SenderKeywords
	}
	if input.SubjectKeywords != nil {
		rule.SubjectKeywords = input.SubjectKeywords
	}
	if input.RequireAttachment != nil {
		rule.RequireAttachment = *input.RequireAttachment
	}

	if err := s.repo.SaveScanRule(rule); err != nil {
		log.WithFields(logger.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to save Gmail scan settings")
		return errors.NewDBError("save Gmail settings", err)
	}

	log.WithField("user_id", userID).Info("Gmail settings updated")
	return nil
}

// GetSettings returns the current Gmail scan settings
func (s *GmailService) GetSettings(userID uint) (*models.GmailScanRule, error) {
	rule, err := s.repo.GetScanRule(userID)
	if err != nil {
		return &models.GmailScanRule{
			UserID:            userID,
			Enabled:           false,
			SenderKeywords:    []string{"credit", "信用卡", "帳單", "statement"},
			SubjectKeywords:   []string{"帳單", "電子帳單", "statement"},
			RequireAttachment: true,
		}, nil
	}
	return rule, nil
}

// GetOAuthTokenForUser returns a valid OAuth2 token for the user, refreshing if needed
func (s *GmailService) GetOAuthTokenForUser(userID uint) (*oauth2.Token, error) {
	token, err := s.repo.GetToken(userID)
	if err != nil {
		return nil, errors.NewNotFoundError("Gmail connection", userID)
	}

	oauthToken, err := s.getOAuthToken(token)
	if err != nil {
		return nil, err
	}

	// If the token was refreshed, save the new one
	if oauthToken.Expiry.After(time.Now()) && (token.TokenExpiry == nil || oauthToken.Expiry.After(*token.TokenExpiry)) {
		accessTokenEnc, encErr := s.crypto.Encrypt(oauthToken.AccessToken)
		if encErr == nil {
			token.AccessTokenEncrypted = accessTokenEnc
			token.TokenExpiry = &oauthToken.Expiry
			_ = s.repo.SaveToken(token)
		}
	}

	return oauthToken, nil
}

// --- Internal helpers ---

func (s *GmailService) getOAuthToken(token *models.GmailToken) (*oauth2.Token, error) {
	accessToken, err := s.crypto.Decrypt(token.AccessTokenEncrypted)
	if err != nil {
		return nil, errors.NewDecryptionError(err)
	}

	refreshToken, err := s.crypto.Decrypt(token.RefreshTokenEncrypted)
	if err != nil {
		return nil, errors.NewDecryptionError(err)
	}

	oauthToken := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}
	if token.TokenExpiry != nil {
		oauthToken.Expiry = *token.TokenExpiry
	}

	// Auto-refresh if expired
	tokenSource := s.oauthConfig.TokenSource(context.Background(), oauthToken)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, errors.NewInternalError("Failed to refresh Gmail token", err)
	}

	return newToken, nil
}

func (s *GmailService) fetchGmailEmail(token *oauth2.Token) (string, error) {
	ctx := context.Background()
	svc, err := gmail.NewService(ctx, option.WithTokenSource(s.oauthConfig.TokenSource(ctx, token)))
	if err != nil {
		return "", err
	}

	profile, err := svc.Users.GetProfile("me").Do()
	if err != nil {
		return "", err
	}

	return profile.EmailAddress, nil
}

func (s *GmailService) revokeToken(accessToken string) {
	log := logger.ServiceLog("GmailService", "revokeToken")
	// Best-effort revocation via Google's revoke endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})).
		Post("https://oauth2.googleapis.com/revoke?token="+accessToken, "application/x-www-form-urlencoded", nil)
	if err != nil {
		log.WithError(err).Warn("Failed to revoke Gmail token (best-effort)")
	}
}

func (s *GmailService) generateState(userID uint) string {
	data := fmt.Sprintf("%d:%d", userID, time.Now().Unix())
	mac := hmac.New(sha256.New, []byte(s.stateSecret))
	mac.Write([]byte(data))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%d:%d:%s", userID, time.Now().Unix(), sig)
}

func (s *GmailService) verifyState(state string, userID uint) bool {
	var stateUserID uint
	var timestamp int64
	var sig string

	n, err := fmt.Sscanf(state, "%d:%d:%s", &stateUserID, &timestamp, &sig)
	if err != nil || n != 3 {
		return false
	}

	// Verify user ID matches
	if stateUserID != userID {
		return false
	}

	// Verify state is not expired (10 minutes)
	if time.Now().Unix()-timestamp > 600 {
		return false
	}

	// Verify HMAC signature
	data := fmt.Sprintf("%d:%d", stateUserID, timestamp)
	mac := hmac.New(sha256.New, []byte(s.stateSecret))
	mac.Write([]byte(data))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(sig), []byte(expectedSig))
}
