package service

import (
	"fmt"
	"time"

	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// PeerOnlineThreshold is the number of seconds after which a peer is considered offline
// if no heartbeat has been received.
const PeerOnlineThreshold = 90

// FindOrCreatePeer finds a peer by peerID and UUID, or creates a new one.
// Returns the peer and whether it was newly created.
func FindOrCreatePeer(peerID, uuid string) (*model.Peer, bool, error) {
	var peer model.Peer
	err := database.DB.Where("peer_id = ? AND uuid = ?", peerID, uuid).First(&peer).Error
	if err == nil {
		return &peer, false, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, false, fmt.Errorf("failed to query peer: %w", err)
	}

	// Create new peer
	peer = model.Peer{
		PeerID:   peerID,
		UUID:     uuid,
		IsOnline: false,
	}

	if err := database.DB.Create(&peer).Error; err != nil {
		return nil, false, fmt.Errorf("failed to create peer: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"peer_id": peerID,
		"uuid":    uuid,
	}).Info("new peer registered")

	return &peer, true, nil
}

// UpdatePeerHeartbeat updates the peer's last seen time, IP, online status, and version.
// Returns whether the peer was previously offline (used for WebSocket status push).
func UpdatePeerHeartbeat(peer *model.Peer, ip, version string) (wasOffline bool) {
	wasOffline = !peer.IsOnline

	now := time.Now().Unix()
	updates := map[string]any{
		"last_online_time": now,
		"last_online_ip":   ip,
		"is_online":        true,
	}

	if version != "" {
		updates["version"] = version
	}

	database.DB.Model(peer).Updates(updates)

	// Update local struct for caller use
	peer.IsOnline = true
	peer.LastOnlineTime = now
	peer.LastOnlineIP = ip
	if version != "" {
		peer.Version = version
	}

	return wasOffline
}

// UpdatePeerSysinfo updates the peer's system information (cpu, memory, os, hostname, username, version).
func UpdatePeerSysinfo(peer *model.Peer, cpu, memory, osName, hostname, username, version string) error {
	updates := map[string]any{
		"cpu":      cpu,
		"memory":   memory,
		"os":       osName,
		"hostname": hostname,
		"username": username,
		"version":  version,
	}

	if err := database.DB.Model(peer).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update peer sysinfo: %w", err)
	}

	// Update local struct
	peer.CPU = cpu
	peer.Memory = memory
	peer.OS = osName
	peer.Hostname = hostname
	peer.Username = username
	peer.Version = version

	return nil
}

// CheckOfflinePeers scans for peers that are marked online but have not sent
// a heartbeat within the threshold. It updates their status and returns the
// list of peer IDs that went offline.
//
// Uses a two-phase approach to avoid race conditions:
//  1. Collect peer IDs from a snapshot query (for broadcasting)
//  2. Do an atomic batch UPDATE with the same WHERE clause — if a heartbeat
//     arrives between steps 1 and 2, last_online_time is bumped and the row
//     is excluded from the UPDATE. The broadcast may contain stale "offline"
//     events, but the next heartbeat will emit a corrective "online" event.
func CheckOfflinePeers() []string {
	threshold := time.Now().Unix() - PeerOnlineThreshold

	// Phase 1: snapshot of stale peer IDs for WebSocket broadcast
	var stalePeers []model.Peer
	database.DB.Where("is_online = ? AND last_online_time < ?", true, threshold).Find(&stalePeers)

	offlineIDs := make([]string, 0, len(stalePeers))
	for _, p := range stalePeers {
		offlineIDs = append(offlineIDs, p.PeerID)
	}

	// Phase 2: atomic batch update — only truly stale rows are affected
	result := database.DB.Model(&model.Peer{}).
		Where("is_online = ? AND last_online_time < ?", true, threshold).
		Update("is_online", false)

	if result.Error != nil {
		logrus.WithError(result.Error).Error("failed to mark offline peers")
		return offlineIDs
	}

	if result.RowsAffected > 0 {
		logrus.WithFields(logrus.Fields{
			"count": result.RowsAffected,
			"stale": len(stalePeers),
		}).Info("peers marked offline (timeout)")
	}

	return offlineIDs
}

// FindPeerByID looks up a peer by its primary key.
func FindPeerByID(id uint) (*model.Peer, error) {
	var peer model.Peer
	if err := database.DB.First(&peer, id).Error; err != nil {
		return nil, fmt.Errorf("peer not found: %w", err)
	}
	return &peer, nil
}

// FindPeerByPeerIDAndUUID looks up a peer by device ID and hardware UUID.
func FindPeerByPeerIDAndUUID(peerID, uuid string) (*model.Peer, error) {
	var peer model.Peer
	if err := database.DB.Where("peer_id = ? AND uuid = ?", peerID, uuid).First(&peer).Error; err != nil {
		return nil, fmt.Errorf("peer not found: %w", err)
	}
	return &peer, nil
}

// DeletePeersByIDs batch-deletes peers by their primary key IDs.
func DeletePeersByIDs(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	result := database.DB.Where("id IN ?", ids).Delete(&model.Peer{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete peers: %w", result.Error)
	}

	logrus.WithField("count", result.RowsAffected).Info("peers deleted")
	return nil
}
