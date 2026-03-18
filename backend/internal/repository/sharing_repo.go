package repository

import (
	"billing-note/internal/models"
	"errors"

	"gorm.io/gorm"
)

type SharingRepository interface {
	GetPairingCode(userID uint) (*models.UserPairingCode, error)
	SavePairingCode(code *models.UserPairingCode) error
	FindByCode(code string) (*models.UserPairingCode, error)
	CreateSharedAccess(access *models.SharedAccess) error
	ListSharedByOwner(ownerID uint) ([]models.SharedAccess, error)
	ListSharedByViewer(viewerID uint) ([]models.SharedAccess, error)
	DeleteSharedAccess(ownerID, viewerID uint) error
	HasAccess(ownerID, viewerID uint) (bool, error)
}

type sharingRepository struct {
	db *gorm.DB
}

func NewSharingRepository(db *gorm.DB) SharingRepository {
	return &sharingRepository{db: db}
}

func (r *sharingRepository) GetPairingCode(userID uint) (*models.UserPairingCode, error) {
	var code models.UserPairingCode
	err := r.db.Where("user_id = ?", userID).First(&code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &code, nil
}

func (r *sharingRepository) SavePairingCode(code *models.UserPairingCode) error {
	return r.db.Save(code).Error
}

func (r *sharingRepository) FindByCode(code string) (*models.UserPairingCode, error) {
	var pairingCode models.UserPairingCode
	err := r.db.Where("code = ?", code).First(&pairingCode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &pairingCode, nil
}

func (r *sharingRepository) CreateSharedAccess(access *models.SharedAccess) error {
	return r.db.Create(access).Error
}

func (r *sharingRepository) ListSharedByOwner(ownerID uint) ([]models.SharedAccess, error) {
	var accesses []models.SharedAccess
	err := r.db.Preload("Viewer").Where("owner_id = ?", ownerID).Find(&accesses).Error
	return accesses, err
}

func (r *sharingRepository) ListSharedByViewer(viewerID uint) ([]models.SharedAccess, error) {
	var accesses []models.SharedAccess
	err := r.db.Preload("Owner").Where("viewer_id = ?", viewerID).Find(&accesses).Error
	return accesses, err
}

func (r *sharingRepository) DeleteSharedAccess(ownerID, viewerID uint) error {
	return r.db.Where("owner_id = ? AND viewer_id = ?", ownerID, viewerID).Delete(&models.SharedAccess{}).Error
}

func (r *sharingRepository) HasAccess(ownerID, viewerID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.SharedAccess{}).Where("owner_id = ? AND viewer_id = ?", ownerID, viewerID).Count(&count).Error
	return count > 0, err
}
