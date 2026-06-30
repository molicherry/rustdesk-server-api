package service

import (
	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
)

// ListAuditConns returns paginated connection audit logs.
// When orgID > 0, filters to logs within that organization.
func ListAuditConns(orgID uint, page, pageSize int, peerID string, startTime, endTime int64) ([]model.AuditConn, int64, error) {
	query := database.DB.Model(&model.AuditConn{})

	if orgID > 0 {
		query = query.Where("organization_id = ?", orgID)
	}
	if peerID != "" {
		query = query.Where("peer_id = ? OR from_peer = ?", peerID, peerID)
	}
	if startTime > 0 {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("created_at <= ?", endTime)
	}

	var total int64
	query.Count(&total)

	var logs []model.AuditConn
	offset := (page - 1) * pageSize
	query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs)

	return logs, total, nil
}

// ListAuditFiles returns paginated file transfer audit logs.
// When orgID > 0, filters to logs within that organization.
func ListAuditFiles(orgID uint, page, pageSize int, peerID string, startTime, endTime int64) ([]model.AuditFile, int64, error) {
	query := database.DB.Model(&model.AuditFile{})

	if orgID > 0 {
		query = query.Where("organization_id = ?", orgID)
	}
	if peerID != "" {
		query = query.Where("peer_id = ? OR from_peer = ?", peerID, peerID)
	}
	if startTime > 0 {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("created_at <= ?", endTime)
	}

	var total int64
	query.Count(&total)

	var logs []model.AuditFile
	offset := (page - 1) * pageSize
	query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs)

	return logs, total, nil
}

// ListLoginLogs returns paginated login logs.
// When orgID > 0, filters to logs within that organization.
func ListLoginLogs(orgID uint, page, pageSize int, userID uint) ([]model.LoginLog, int64, error) {
	query := database.DB.Model(&model.LoginLog{})

	if orgID > 0 {
		query = query.Where("organization_id = ?", orgID)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	var total int64
	query.Count(&total)

	var logs []model.LoginLog
	offset := (page - 1) * pageSize
	query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs)

	return logs, total, nil
}
