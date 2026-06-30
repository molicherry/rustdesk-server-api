package model

import (
	"time"

	"gorm.io/gorm"
)

// TfaSetupToken stores a TOTP secret during the TFA enable flow.
// The secret is stored temporarily until the user verifies the code;
// on successful verification it is moved to user.TFASecret and this record is deleted.
type TfaSetupToken struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	Secret    string         `gorm:"not null;size:100" json:"-"`
	QrURL     string         `gorm:"not null;size:512" json:"qr_url"`
	ExpiresAt int64          `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for TfaSetupToken.
func (TfaSetupToken) TableName() string {
	return "tfa_setup_tokens"
}
