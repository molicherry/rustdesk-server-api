package model

import (
	"time"

	"gorm.io/gorm"
)

// LoginLog records user login events for auditing.
type LoginLog struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	Client    string         `gorm:"size:100" json:"client"`
	DeviceID  string         `gorm:"size:100" json:"device_id"`
	UUID      string         `gorm:"size:255" json:"uuid"`
	IP        string         `gorm:"size:50" json:"ip"`
	Type      string         `gorm:"size:50" json:"type"`
	Platform  string         `gorm:"size:50" json:"platform"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for LoginLog.
func (LoginLog) TableName() string {
	return "login_logs"
}
