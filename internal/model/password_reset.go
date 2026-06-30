package model

import (
	"time"

	"gorm.io/gorm"
)

// PasswordReset stores a reset token for the forgot-password flow.
type PasswordReset struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Email     string         `gorm:"index;not null;size:255" json:"email"`
	Token     string         `gorm:"uniqueIndex;not null;size:64" json:"token"`
	Code      string         `gorm:"not null;size:10" json:"code"`
	ExpiresAt int64          `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for PasswordReset.
func (PasswordReset) TableName() string {
	return "password_resets"
}
