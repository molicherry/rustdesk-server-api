package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/middleware"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
)

// PeerResponse is the public peer representation returned in API responses.
type PeerResponse struct {
	ID             uint   `json:"id"`
	PeerID         string `json:"peer_id"`
	UUID           string `json:"uuid"`
	Hostname       string `json:"hostname"`
	OS             string `json:"os"`
	Username       string `json:"username"`
	Version        string `json:"version"`
	CPU            string `json:"cpu"`
	Memory         string `json:"memory"`
	LastOnlineTime int64  `json:"last_online_time"`
	LastOnlineIP   string `json:"last_online_ip"`
	Alias          string `json:"alias"`
	IsOnline       bool   `json:"is_online"`
	Note           string `json:"note"`
}

// peerResponse converts a model.Peer to the API response struct.
func peerResponse(p *model.Peer) PeerResponse {
	return PeerResponse{
		ID:             p.ID,
		PeerID:         p.PeerID,
		UUID:           p.UUID,
		Hostname:       p.Hostname,
		OS:             p.OS,
		Username:       p.Username,
		Version:        p.Version,
		CPU:            p.CPU,
		Memory:         p.Memory,
		LastOnlineTime: p.LastOnlineTime,
		LastOnlineIP:   p.LastOnlineIP,
		Alias:          p.Alias,
		IsOnline:       p.IsOnline,
		Note:           p.Note,
	}
}

// PeerListRequest holds query parameters for the peer list endpoint.
type PeerListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"` // "online", "offline", "all"
	OS       string `form:"os"`
	Search   string `form:"search"` // search in peer_id, hostname, alias
	OrgID    uint   `form:"org_id"`
}

// PeerListResponse is the response body for the peer list endpoint.
type PeerListResponse struct {
	Total int64          `json:"total"`
	Data  []PeerResponse `json:"data"`
}

// ListPeers handles GET /api/admin/peer/list
func ListPeers(c *gin.Context) {
	var req PeerListRequest
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

	// Get org_id from middleware context (set by RequireOrgRole)
	orgID, _ := c.Get(middleware.ContextKeyOrgID)
	var oid uint
	if orgID != nil {
		oid = orgID.(uint)
	}
	// Override with query param if provided (admin global view can specify)
	if req.OrgID > 0 {
		oid = req.OrgID
	}

	peers, total, err := service.ListPeers(oid, req.Page, req.PageSize, req.Status, req.OS, req.Search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "internal server error",
		})
		return
	}

	data := make([]PeerResponse, len(peers))
	for i, p := range peers {
		data[i] = peerResponse(&p)
	}

	c.JSON(http.StatusOK, PeerListResponse{
		Total: total,
		Data:  data,
	})
}

// GetPeer handles GET /api/admin/peer/detail/:id
func GetPeer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid peer ID",
		})
		return
	}

	peer, err := service.FindPeerByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Peer not found",
		})
		return
	}

	c.JSON(http.StatusOK, peerResponse(peer))
}

// DeletePeerRequest is the request body for deleting peers.
type DeletePeerRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// DeletePeer handles POST /api/admin/peer/delete
func DeletePeer(c *gin.Context) {
	var req DeletePeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "ids is required (array of peer IDs)",
		})
		return
	}

	if err := service.DeletePeersByIDs(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to delete peers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Peers deleted successfully"})
}

// BatchDeletePeer handles POST /api/admin/peer/batchDelete
func BatchDeletePeer(c *gin.Context) {
	var req DeletePeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "ids is required (array of peer IDs)",
		})
		return
	}

	if err := service.DeletePeersByIDs(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to delete peers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Peers deleted successfully"})
}

// UpdatePeerRequest is the request body for updating a peer.
type UpdatePeerRequest struct {
	ID            uint   `json:"id" binding:"required"`
	Alias         string `json:"alias"`
	Note          string `json:"note"`
	DeviceGroupID *uint  `json:"device_group_id"`
}

// UpdatePeer handles POST /api/admin/peer/update
func UpdatePeer(c *gin.Context) {
	var req UpdatePeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "id is required",
		})
		return
	}

	peer, err := service.FindPeerByID(req.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Peer not found",
		})
		return
	}

	updates := map[string]any{}
	if req.Alias != "" {
		updates["alias"] = req.Alias
	}
	if req.Note != "" {
		updates["note"] = req.Note
	}
	if req.DeviceGroupID != nil {
		updates["device_group_id"] = *req.DeviceGroupID
	}

	if len(updates) > 0 {
		if err := database.DB.Model(peer).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "server_error",
				"message": "Failed to update peer",
			})
			return
		}
		// Refresh from DB
		database.DB.First(peer, peer.ID)
	}

	c.JSON(http.StatusOK, peerResponse(peer))
}
