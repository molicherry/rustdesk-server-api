package model

import (
	"time"

	"gorm.io/gorm"
)

// AddressBook stores a user's address book contacts for cross-device sync.
type AddressBook struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	UserID           uint           `gorm:"index;not null" json:"user_id"`
	PeerID           string         `gorm:"index;size:100" json:"peer_id"`
	Username         string         `gorm:"size:100" json:"username"`
	Hostname         string         `gorm:"size:255" json:"hostname"`
	Alias            string         `gorm:"size:255" json:"alias"`
	Platform         string         `gorm:"size:50" json:"platform"`
	Tags             string         `gorm:"type:text" json:"tags"` // JSON array string
	ForceAlwaysRelay bool           `gorm:"default:false" json:"force_always_relay"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for AddressBook.
func (AddressBook) TableName() string {
	return "address_books"
}
