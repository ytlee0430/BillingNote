package repository

import (
	"billing-note/internal/models"

	"gorm.io/gorm"
)

type CategoryKeywordRepository interface {
	ListByUser(userID uint) ([]models.CategoryKeyword, error)
	Create(kw *models.CategoryKeyword) error
	Delete(id, userID uint) error
	DeleteByUserAndCategory(userID, categoryID uint) error
	BatchCreate(keywords []models.CategoryKeyword) error
}

type categoryKeywordRepository struct {
	db *gorm.DB
}

func NewCategoryKeywordRepository(db *gorm.DB) CategoryKeywordRepository {
	return &categoryKeywordRepository{db: db}
}

func (r *categoryKeywordRepository) ListByUser(userID uint) ([]models.CategoryKeyword, error) {
	var keywords []models.CategoryKeyword
	err := r.db.Preload("Category").Where("user_id = ?", userID).Order("category_id, keyword").Find(&keywords).Error
	return keywords, err
}

func (r *categoryKeywordRepository) Create(kw *models.CategoryKeyword) error {
	return r.db.Create(kw).Error
}

func (r *categoryKeywordRepository) Delete(id, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.CategoryKeyword{}).Error
}

func (r *categoryKeywordRepository) DeleteByUserAndCategory(userID, categoryID uint) error {
	return r.db.Where("user_id = ? AND category_id = ?", userID, categoryID).Delete(&models.CategoryKeyword{}).Error
}

func (r *categoryKeywordRepository) BatchCreate(keywords []models.CategoryKeyword) error {
	if len(keywords) == 0 {
		return nil
	}
	return r.db.Create(&keywords).Error
}
