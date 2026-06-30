package model

import (
	"time"

	"gorm.io/gorm"
)

// Organization represents a tenant/organization for multi-tenancy.
type Organization struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null;size:255" json:"name"`
	Description string         `gorm:"size:1000" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Organization.
func (Organization) TableName() string {
	return "organizations"
}
