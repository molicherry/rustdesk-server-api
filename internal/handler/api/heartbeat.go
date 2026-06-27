package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
	"github.com/rustdesk/rustdesk-api-server/internal/wsutil"
	"github.com/sirupsen/logrus"
)

// HeartbeatRequest is the request body from the RustDesk client for /api/heartbeat.
// MUST match the client protocol exactly.
type HeartbeatRequest struct {
	ID         string `json:"id"`
	UUID       string `json:"uuid"`
	Ver        int64  `json:"ver"`
	Conns      []int  `json:"conns"`
	ModifiedAt int64  `json:"modified_at"`
}

// HeartbeatResponse is the response for /api/heartbeat.
// MUST match what RustDesk clients expect.
type HeartbeatResponse struct {
	Sysinfo    *string        `json:"sysinfo,omitempty"`
	Disconnect []int          `json:"disconnect"`
	ModifiedAt int64          `json:"modified_at"`
	Strategy   map[string]any `json:"strategy"`
}

// Heartbeat handles POST /api/heartbeat
// RustDesk clients send device heartbeat periodically to update online status.
// The response may request sysinfo upload and/or connection disconnects.
func Heartbeat(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid request body",
		})
		return
	}

	if req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "id is required",
		})
		return
	}

	// Find existing peer or create new one
	peer, isNew, err := service.FindOrCreatePeer(req.ID, req.UUID)
	if err != nil {
		logrus.WithError(err).Error("failed to find or create peer")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to process heartbeat",
		})
		return
	}

	// Update heartbeat info
	wasOffline := service.UpdatePeerHeartbeat(peer, c.ClientIP(), "")

	// Determine if sysinfo should be requested
	// Request sysinfo for: new peers, version changes (when ver > 0)
	needsSysinfo := isNew
	if req.Ver > 0 && peer.Version == "" {
		needsSysinfo = true
	}

	resp := HeartbeatResponse{
		Disconnect: []int{},
		ModifiedAt: time.Now().Unix(),
		Strategy:   map[string]any{},
	}

	if needsSysinfo {
		reqStr := "REQUESTED"
		resp.Sysinfo = &reqStr
	}

	// WebSocket broadcast on state change
	if wasOffline && wsutil.GlobalHub != nil {
		wsutil.GlobalHub.BroadcastToAuthenticated("peer_online", gin.H{
			"id":       peer.PeerID,
			"hostname": peer.Hostname,
			"os":       peer.OS,
			"time":     time.Now().Unix(),
		})
	}

	c.JSON(http.StatusOK, resp)
}
