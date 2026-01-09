package services

import (
	"billing-note/internal/models"
	"billing-note/pkg/crypto"
	"errors"
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
	aesCrypto, err := crypto.NewAESCrypto(encryptionKey)
	if err != nil {
		return nil, err
	}

	return &PDFPasswordService{
		db:     db,
		crypto: aesCrypto,
	}, nil
}

// SetPassword sets or updates a PDF password for a user
func (s *PDFPasswordService) SetPassword(userID uint, input models.PDFPasswordInput) error {
	// Encrypt the password
	encrypted, err := s.crypto.Encrypt(input.Password)
	if err != nil {
		return err
	}

	// Check if password with this priority already exists
	var existing models.PDFPassword
	err = s.db.Where("user_id = ? AND priority = ?", userID, input.Priority).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new password
		newPassword := models.PDFPassword{
			UserID:            userID,
			PasswordEncrypted: encrypted,
			Priority:          input.Priority,
			Label:             input.Label,
		}
		return s.db.Create(&newPassword).Error
	}

	if err != nil {
		return err
	}

	// Update existing password
	existing.PasswordEncrypted = encrypted
	existing.Label = input.Label
	return s.db.Save(&existing).Error
}

// GetPasswords returns all PDF passwords for a user (without actual passwords)
func (s *PDFPasswordService) GetPasswords(userID uint) ([]models.PDFPasswordResponse, error) {
	var passwords []models.PDFPassword
	err := s.db.Where("user_id = ?", userID).Order("priority ASC").Find(&passwords).Error
	if err != nil {
		return nil, err
	}

	responses := make([]models.PDFPasswordResponse, len(passwords))
	for i, p := range passwords {
		responses[i] = p.ToResponse()
	}

	return responses, nil
}

// GetDecryptedPasswords returns decrypted passwords for PDF parsing
func (s *PDFPasswordService) GetDecryptedPasswords(userID uint) ([]string, error) {
	var passwords []models.PDFPassword
	err := s.db.Where("user_id = ?", userID).Order("priority ASC").Find(&passwords).Error
	if err != nil {
		return nil, err
	}

	decrypted := make([]string, 0, len(passwords))
	for _, p := range passwords {
		if p.PasswordEncrypted == "" {
			continue
		}
		plaintext, err := s.crypto.Decrypt(p.PasswordEncrypted)
		if err != nil {
			continue // Skip passwords that can't be decrypted
		}
		decrypted = append(decrypted, plaintext)
	}

	return decrypted, nil
}

// DeletePassword deletes a PDF password
func (s *PDFPasswordService) DeletePassword(userID uint, priority int) error {
	result := s.db.Where("user_id = ? AND priority = ?", userID, priority).Delete(&models.PDFPassword{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("password not found")
	}
	return nil
}

// SetMultiplePasswords sets multiple passwords at once
func (s *PDFPasswordService) SetMultiplePasswords(userID uint, inputs []models.PDFPasswordInput) error {
	// Sort by priority
	sort.Slice(inputs, func(i, j int) bool {
		return inputs[i].Priority < inputs[j].Priority
	})

	for _, input := range inputs {
		if input.Password == "" {
			continue
		}
		if err := s.SetPassword(userID, input); err != nil {
			return err
		}
	}

	return nil
}
