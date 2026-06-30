package model

import (
	"time"

	"gorm.io/gorm"
)

// DeviceGroup organizes peers into logical groups.
type DeviceGroup struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name           string         `gorm:"not null;size:255" json:"name"`
	OrganizationID uint           `gorm:"default:0" json:"organization_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for DeviceGroup.
func (DeviceGroup) TableName() string {
	return "device_groups"
}
