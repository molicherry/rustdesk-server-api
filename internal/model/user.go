package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user account in the system.
type User struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Username      string         `gorm:"uniqueIndex;not null;size:100" json:"username"`
	Email         string         `gorm:"size:255" json:"email"`
	Password      string         `gorm:"not null;size:255" json:"-"`
	Nickname      string         `gorm:"size:100" json:"nickname"`
	Avatar        string         `gorm:"size:500" json:"avatar"`
	IsAdmin       bool           `gorm:"default:false" json:"is_admin"`
	Role          string         `gorm:"default:admin;size:50" json:"role"` // admin, user, auditor
	Status        int            `gorm:"default:1" json:"status"` // 1=active, 0=disabled
	TFASecret     string         `gorm:"size:100" json:"-"`
	EmailVerified bool           `gorm:"default:false" json:"email_verified"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for User.
func (User) TableName() string {
	return "users"
}
