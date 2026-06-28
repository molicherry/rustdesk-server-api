package model

import (
	"time"

	"gorm.io/gorm"
)

// Peer represents a RustDesk client device.
// Populated via /api/heartbeat and /api/sysinfo endpoints.
type Peer struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	PeerID         string         `gorm:"uniqueIndex:idx_peer_id_uuid;size:100" json:"peer_id"`
	UUID           string         `gorm:"uniqueIndex:idx_peer_id_uuid;size:255" json:"uuid"`
	Hostname       string         `gorm:"size:255" json:"hostname"`
	OS             string         `gorm:"size:100" json:"os"`
	Username       string         `gorm:"size:100" json:"username"`
	Version        string         `gorm:"size:50" json:"version"`
	CPU            string         `gorm:"size:255" json:"cpu"`
	Memory         string         `gorm:"size:100" json:"memory"`
	LastOnlineTime int64          `json:"last_online_time"`
	LastOnlineIP   string         `gorm:"size:50" json:"last_online_ip"`
	Alias          string         `gorm:"size:255" json:"alias"`
	IsOnline       bool           `gorm:"default:false" json:"is_online"`
	Note           string         `gorm:"size:500" json:"note"`
	DeviceGroupID  *uint          `json:"device_group_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Peer.
func (Peer) TableName() string {
	return "peers"
}
