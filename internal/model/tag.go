package model

import (
	"time"

	"gorm.io/gorm"
)

// Tag represents a user-created label for organizing address book contacts.
type Tag struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"uniqueIndex:idx_user_tag_name;not null" json:"user_id"`
	Name      string         `gorm:"uniqueIndex:idx_user_tag_name;not null;size:100" json:"name"`
	Color          int64          `json:"color"` // Flutter ARGB color value
	OrganizationID uint           `gorm:"default:0" json:"organization_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Tag.
func (Tag) TableName() string {
	return "tags"
}
