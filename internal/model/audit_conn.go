package model

import (
	"time"

	"gorm.io/gorm"
)

// AuditConn records remote connection events for auditing.
type AuditConn struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ConnID    int64          `gorm:"index" json:"conn_id"`
	PeerID    string         `gorm:"index;size:100" json:"peer_id"`
	FromPeer  string         `gorm:"size:100" json:"from_peer"`
	FromName  string         `gorm:"size:255" json:"from_name"`
	IP        string         `gorm:"size:50" json:"ip"`
	SessionID string         `gorm:"size:255" json:"session_id"`
	Type      int            `json:"type"`
	UUID      string         `gorm:"size:255" json:"uuid"`
	CloseTime      int64          `json:"close_time"`
	OrganizationID uint           `gorm:"default:0" json:"organization_id"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for AuditConn.
func (AuditConn) TableName() string {
	return "audit_conns"
}
