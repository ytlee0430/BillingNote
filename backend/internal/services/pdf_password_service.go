package services

import (
	"billing-note/internal/models"
	"billing-note/pkg/crypto"
	"billing-note/pkg/errors"
	"billing-note/pkg/logger"
	"sort"

	"gorm.io/gorm"
)

// PDFPasswordService handles PDF password operations
type PDFPasswordService struct {
	db     *gorm.DB
	crypto *crypto.AESCrypto
}

// NewPDFPasswordService creates a new PDF password service
func NewPDFPasswordService(db *gorm.DB, encryptionKey string) (*PDFPasswordService, error) {
	log := logger.ServiceLog("PDFPasswordService", "NewPDFPasswordService")

	log.Debug("Initializing PDF password service with AES encryption")

	aesCrypto, err := crypto.NewAESCrypto(encryptionKey)
	if err != nil {
		log.WithError(err).Error("Failed to initialize AES encryption")
		return nil, errors.NewEncryptionError("initialization", err)
	}

	log.Info("PDF password service initialized successfully")

	return &PDFPasswordService{
		db:     db,
		crypto: aesCrypto,
	}, nil
}

// SetPassword sets or updates a PDF password for a user
func (s *PDFPasswordService) SetPassword(userID uint, input models.PDFPasswordInput) error {
	log := logger.ServiceLog("PDFPasswordService", "SetPassword")

	log.WithFields(logger.Fields{
		"user_id":  userID,
		"priority": input.Priority,
		"label":    input.Label,
	}).Debug("Setting PDF password")

	// Encrypt the password
	encrypted, err := s.crypto.Encrypt(input.Password)
	if err != nil {
		log.WithFields(logger.Fields{
			"user_id":  userID,
			"priority": input.Priority,
			"error":    err.Error(),
		}).Error("Failed to encrypt password")
		return errors.NewEncryptionError("password encryption", err)
	}

	// Check if password with this priority already exists
	var existing models.PDFPassword
	err = s.db.Where("user_id = ? AND priority = ?", userID, input.Priority).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new password
		log.WithFields(logger.Fields{
			"user_id":  userID,
			"priority": input.Priority,
		}).Debug("Creating new PDF password entry")

		newPassword := models.PDFPassword{
			UserID:            userID,
			PasswordEncrypted: encrypted,
			Priority:          input.Priority,
			Label:             input.Label,
		}
		if err := s.db.Create(&newPassword).Error; err != nil {
			log.WithFields(logger.Fields{
				"user_id":  userID,
				"priority": input.Priority,
				"error":    err.Error(),
			}).Error("Failed to create PDF password in database")
			return errors.NewDBError("create PDF password", err)
		}

		log.WithFields(logger.Fields{
			"user_id":  userID,
			"priority": input.Priority,
		}).Info("PDF password created successfully")
		return nil
	}

	if err != nil {
		log.WithFields(logger.Fields{
			"user_id":  userID,
			"priority": input.Priority,
			"error":    err.Error(),
		}).Error("Failed to check existing PDF password")
		return errors.NewDBError("check existing PDF password", err)
	}

	// Update existing password
	log.WithFields(logger.Fields{
		"user_id":  userID,
		"priority": input.Priority,
	}).Debug("Updating existing PDF password entry")

	existing.PasswordEncrypted = encrypted
	existing.Label = input.Label
	if err := s.db.Save(&existing).Error; err != nil {
		log.WithFields(logger.Fields{
			"user_id":  userID,
			"priority": input.Priority,
			"error":    err.Error(),
		}).Error("Failed to update PDF password in database")
		return errors.NewDBError("update PDF password", err)
	}

	log.WithFields(logger.Fields{
		"user_id":  userID,
		"priority": input.Priority,
	}).Info("PDF password updated successfully")
	return nil
}

// GetPasswords returns all PDF passwords for a user (without actual passwords)
func (s *PDFPasswordService) GetPasswords(userID uint) ([]models.PDFPasswordResponse, error) {
	log := logger.ServiceLog("PDFPasswordService", "GetPasswords")

	log.WithField("user_id", userID).Debug("Fetching PDF passwords for user")

	var passwords []models.PDFPassword
	err := s.db.Where("user_id = ?", userID).Order("priority ASC").Find(&passwords).Error
	if err != nil {
		log.WithFields(logger.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to fetch PDF passwords from database")
		return nil, errors.NewDBError("fetch PDF passwords", err)
	}

	responses := make([]models.PDFPasswordResponse, len(passwords))
	for i, p := range passwords {
		responses[i] = p.ToResponse()
	}

	log.WithFields(logger.Fields{
		"user_id": userID,
		"count":   len(passwords),
	}).Debug("PDF passwords fetched successfully")

	return responses, nil
}

// GetDecryptedPasswords returns decrypted passwords for PDF parsing
func (s *PDFPasswordService) GetDecryptedPasswords(userID uint) ([]string, error) {
	log := logger.ServiceLog("PDFPasswordService", "GetDecryptedPasswords")

	log.WithField("user_id", userID).Debug("Fetching and decrypting PDF passwords")

	var passwords []models.PDFPassword
	err := s.db.Where("user_id = ?", userID).Order("priority ASC").Find(&passwords).Error
	if err != nil {
		log.WithFields(logger.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to fetch PDF passwords from database")
		return nil, errors.NewDBError("fetch PDF passwords", err)
	}

	decrypted := make([]string, 0, len(passwords))
	decryptedCount := 0
	failedCount := 0

	for _, p := range passwords {
		if p.PasswordEncrypted == "" {
			continue
		}
		plaintext, err := s.crypto.Decrypt(p.PasswordEncrypted)
		if err != nil {
			log.WithFields(logger.Fields{
				"user_id":  userID,
				"priority": p.Priority,
				"error":    err.Error(),
			}).Warn("Failed to decrypt password, skipping")
			failedCount++
			continue
		}
		decrypted = append(decrypted, plaintext)
		decryptedCount++
	}

	log.WithFields(logger.Fields{
		"user_id":         userID,
		"total":           len(passwords),
		"decrypted_count": decryptedCount,
		"failed_count":    failedCount,
	}).Debug("PDF passwords decrypted")

	return decrypted, nil
}

// DeletePassword deletes a PDF password
func (s *PDFPasswordService) DeletePassword(userID uint, priority int) error {
	log := logger.ServiceLog("PDFPasswordService", "DeletePassword")

	log.WithFields(logger.Fields{
		"user_id":  userID,
		"priority": priority,
	}).Info("Deleting PDF password")

	result := s.db.Where("user_id = ? AND priority = ?", userID, priority).Delete(&models.PDFPassword{})
	if result.Error != nil {
		log.WithFields(logger.Fields{
			"user_id":  userID,
			"priority": priority,
			"error":    result.Error.Error(),
		}).Error("Failed to delete PDF password from database")
		return errors.NewDBError("delete PDF password", result.Error)
	}

	if result.RowsAffected == 0 {
		log.WithFields(logger.Fields{
			"user_id":  userID,
			"priority": priority,
		}).Warn("PDF password not found for deletion")
		return errors.NewNotFoundError("PDF password", priority)
	}

	log.WithFields(logger.Fields{
		"user_id":  userID,
		"priority": priority,
	}).Info("PDF password deleted successfully")
	return nil
}

// SetMultiplePasswords sets multiple passwords at once
func (s *PDFPasswordService) SetMultiplePasswords(userID uint, inputs []models.PDFPasswordInput) error {
	log := logger.ServiceLog("PDFPasswordService", "SetMultiplePasswords")

	log.WithFields(logger.Fields{
		"user_id": userID,
		"count":   len(inputs),
	}).Info("Setting multiple PDF passwords")

	// Sort by priority
	sort.Slice(inputs, func(i, j int) bool {
		return inputs[i].Priority < inputs[j].Priority
	})

	successCount := 0
	for _, input := range inputs {
		if input.Password == "" {
			log.WithFields(logger.Fields{
				"user_id":  userID,
				"priority": input.Priority,
			}).Debug("Skipping empty password")
			continue
		}
		if err := s.SetPassword(userID, input); err != nil {
			log.WithFields(logger.Fields{
				"user_id":  userID,
				"priority": input.Priority,
				"error":    err.Error(),
			}).Error("Failed to set password in batch operation")
			return err
		}
		successCount++
	}

	log.WithFields(logger.Fields{
		"user_id":       userID,
		"total":         len(inputs),
		"success_count": successCount,
	}).Info("Multiple PDF passwords set successfully")

	return nil
}
