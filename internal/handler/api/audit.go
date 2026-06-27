package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/sirupsen/logrus"
)

// AuditConnRequest is the request body from the RustDesk client for POST /api/audit/conn.
type AuditConnRequest struct {
	Action    string `json:"action"`
	ConnID    int64  `json:"conn_id"`
	PeerID    string `json:"peer_id"`
	FromPeer  string `json:"from_peer"`
	FromName  string `json:"from_name"`
	IP        string `json:"ip"`
	SessionID string `json:"session_id"`
	Type      int    `json:"type"`
	UUID      string `json:"uuid"`
}

// AuditConn handles POST /api/audit/conn
// RustDesk clients send connection audit events (start/close).
func AuditConn(c *gin.Context) {
	var req AuditConnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid request body",
		})
		return
	}

	log := &model.AuditConn{
		ConnID:    req.ConnID,
		PeerID:    req.PeerID,
		FromPeer:  req.FromPeer,
		FromName:  req.FromName,
		IP:        c.ClientIP(),
		SessionID: req.SessionID,
		Type:      req.Type,
		UUID:      req.UUID,
	}

	if err := database.DB.Create(log).Error; err != nil {
		logrus.WithError(err).Error("failed to save audit conn")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to save audit log",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// AuditFileRequest is the request body from the RustDesk client for POST /api/audit/file.
type AuditFileRequest struct {
	PeerID   string `json:"peer_id"`
	FromPeer string `json:"from_peer"`
	Path     string `json:"path"`
	IsFile   bool   `json:"is_file"`
	Info     string `json:"info"`
	Type     int    `json:"type"`
	UUID     string `json:"uuid"`
	FromName string `json:"from_name"`
}

// AuditFile handles POST /api/audit/file
// RustDesk clients send file transfer audit events.
func AuditFile(c *gin.Context) {
	var req AuditFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid request body",
		})
		return
	}

	log := &model.AuditFile{
		PeerID:   req.PeerID,
		FromPeer: req.FromPeer,
		Path:     req.Path,
		IsFile:   req.IsFile,
		Info:     req.Info,
		Type:     req.Type,
		UUID:     req.UUID,
		FromName: req.FromName,
	}

	if err := database.DB.Create(log).Error; err != nil {
		logrus.WithError(err).Error("failed to save audit file")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to save audit log",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
