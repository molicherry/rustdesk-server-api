package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
)

// AuditConnListRequest holds query parameters for connection audit log listing.
type AuditConnListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	PeerID    string `form:"peer_id"`
	StartTime int64  `form:"start_time"`
	EndTime   int64  `form:"end_time"`
}

// AuditConnListResponse is the paginated list response for connection audit logs.
type AuditConnListResponse struct {
	Total int64   `json:"total"`
	Data  []gin.H `json:"data"`
}

// ListAuditConns handles GET /api/admin/audit_conn/list
func ListAuditConns(c *gin.Context) {
	var req AuditConnListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	query := database.DB.Model(&model.AuditConn{})

	if req.PeerID != "" {
		query = query.Where("peer_id = ? OR from_peer = ?", req.PeerID, req.PeerID)
	}
	if req.StartTime > 0 {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime > 0 {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	var total int64
	query.Count(&total)

	var logs []model.AuditConn
	offset := (req.Page - 1) * req.PageSize
	query.Order("id DESC").Offset(offset).Limit(req.PageSize).Find(&logs)

	// Build response data, joining with peer hostnames
	data := make([]gin.H, len(logs))
	for i, l := range logs {
		hostname := ""
		var peer model.Peer
		if l.PeerID != "" {
			if err := database.DB.Where("peer_id = ?", l.PeerID).First(&peer).Error; err == nil {
				hostname = peer.Hostname
			}
		}
		data[i] = gin.H{
			"id":         l.ID,
			"conn_id":    l.ConnID,
			"peer_id":    l.PeerID,
			"from_peer":  l.FromPeer,
			"from_name":  l.FromName,
			"ip":         l.IP,
			"session_id": l.SessionID,
			"type":       l.Type,
			"uuid":       l.UUID,
			"close_time": l.CloseTime,
			"hostname":   hostname,
			"created_at": l.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, AuditConnListResponse{
		Total: total,
		Data:  data,
	})
}

// AuditFileListRequest holds query parameters for file transfer audit log listing.
type AuditFileListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	PeerID    string `form:"peer_id"`
	StartTime int64  `form:"start_time"`
	EndTime   int64  `form:"end_time"`
}

// AuditFileListResponse is the paginated list response for file transfer audit logs.
type AuditFileListResponse struct {
	Total int64   `json:"total"`
	Data  []gin.H `json:"data"`
}

// ListAuditFiles handles GET /api/admin/audit_file/list
func ListAuditFiles(c *gin.Context) {
	var req AuditFileListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	query := database.DB.Model(&model.AuditFile{})

	if req.PeerID != "" {
		query = query.Where("peer_id = ? OR from_peer = ?", req.PeerID, req.PeerID)
	}
	if req.StartTime > 0 {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime > 0 {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	var total int64
	query.Count(&total)

	var logs []model.AuditFile
	offset := (req.Page - 1) * req.PageSize
	query.Order("id DESC").Offset(offset).Limit(req.PageSize).Find(&logs)

	data := make([]gin.H, len(logs))
	for i, l := range logs {
		hostname := ""
		var peer model.Peer
		if l.PeerID != "" {
			if err := database.DB.Where("peer_id = ?", l.PeerID).First(&peer).Error; err == nil {
				hostname = peer.Hostname
			}
		}
		data[i] = gin.H{
			"id":         l.ID,
			"peer_id":    l.PeerID,
			"from_peer":  l.FromPeer,
			"from_name":  l.FromName,
			"path":       l.Path,
			"is_file":    l.IsFile,
			"info":       l.Info,
			"type":       l.Type,
			"uuid":       l.UUID,
			"hostname":   hostname,
			"created_at": l.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, AuditFileListResponse{
		Total: total,
		Data:  data,
	})
}

// LoginLogListRequest holds query parameters for login log listing.
type LoginLogListRequest struct {
	Page     int  `form:"page"`
	PageSize int  `form:"page_size"`
	UserID   uint `form:"user_id"`
}

// LoginLogListResponse is the paginated list response for login logs.
type LoginLogListResponse struct {
	Total int64   `json:"total"`
	Data  []gin.H `json:"data"`
}

// ListLoginLogs handles GET /api/admin/login_log/list
func ListLoginLogs(c *gin.Context) {
	var req LoginLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	query := database.DB.Model(&model.LoginLog{})

	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}

	var total int64
	query.Count(&total)

	var logs []model.LoginLog
	offset := (req.Page - 1) * req.PageSize
	query.Order("id DESC").Offset(offset).Limit(req.PageSize).Find(&logs)

	data := make([]gin.H, len(logs))
	for i, l := range logs {
		username := ""
		var user model.User
		if l.UserID > 0 {
			if err := database.DB.First(&user, l.UserID).Error; err == nil {
				username = user.Username
			}
		}
		data[i] = gin.H{
			"id":         l.ID,
			"user_id":    l.UserID,
			"username":   username,
			"client":     l.Client,
			"device_id":  l.DeviceID,
			"uuid":       l.UUID,
			"ip":         l.IP,
			"type":       l.Type,
			"platform":   l.Platform,
			"created_at": l.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, LoginLogListResponse{
		Total: total,
		Data:  data,
	})
}
