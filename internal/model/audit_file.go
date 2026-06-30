package model

import (
	"time"

	"gorm.io/gorm"
)

// AuditFile records file transfer events for auditing.
type AuditFile struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PeerID    string         `gorm:"index;size:100" json:"peer_id"`
	FromPeer  string         `gorm:"index;size:100" json:"from_peer"`
	Path      string         `gorm:"size:1000" json:"path"`
	IsFile    bool           `json:"is_file"`
	Info      string         `gorm:"type:text" json:"info"`
	Type      int            `json:"type"`
	UUID      string         `gorm:"size:255" json:"uuid"`
	FromName       string         `gorm:"size:255" json:"from_name"`
	OrganizationID uint           `gorm:"default:0" json:"organization_id"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for AuditFile.
func (AuditFile) TableName() string {
	return "audit_files"
}
