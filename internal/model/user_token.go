package model

import (
	"time"

	"gorm.io/gorm"
)

// UserToken stores authentication tokens for API access.
type UserToken struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"index;not null" json:"user_id"`
	DeviceUUID string         `gorm:"size:255" json:"device_uuid"`
	Token      string         `gorm:"index;size:64;not null" json:"token"`
	ExpiredAt  int64          `json:"expired_at"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for UserToken.
func (UserToken) TableName() string {
	return "user_tokens"
}
