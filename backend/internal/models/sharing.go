package models

import "time"

type UserPairingCode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex" json:"user_id"`
	Code      string    `gorm:"not null;uniqueIndex;size:9" json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserPairingCode) TableName() string {
	return "user_pairing_codes"
}

type SharedAccess struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OwnerID   uint      `gorm:"not null;index" json:"owner_id"`
	ViewerID  uint      `gorm:"not null;index" json:"viewer_id"`
	CreatedAt time.Time `json:"created_at"`

	Owner  User `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Viewer User `gorm:"foreignKey:ViewerID" json:"viewer,omitempty"`
}

func (SharedAccess) TableName() string {
	return "shared_access"
}
