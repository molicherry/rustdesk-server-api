package model

import (
	"time"

	"gorm.io/gorm"
)

// EmailVerification stores a 6-digit verification code sent via email.
// Used for email ownership verification and password reset flows.
type EmailVerification struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Email     string         `gorm:"index;not null;size:255" json:"email"`
	Code      string         `gorm:"not null;size:10" json:"code"`
	ExpiresAt int64          `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for EmailVerification.
func (EmailVerification) TableName() string {
	return "email_verifications"
}
