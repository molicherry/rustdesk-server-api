package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
	"github.com/sirupsen/logrus"
)

// SysinfoRequest is the request body from the RustDesk client for /api/sysinfo.
type SysinfoRequest struct {
	CPU      string `json:"cpu"`
	Memory   string `json:"memory"`
	OS       string `json:"os"`
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	Version  string `json:"version"`
	ID       string `json:"id"`
	UUID     string `json:"uuid"`
}

// Sysinfo handles POST /api/sysinfo
// CRITICAL: Response must be PLAIN TEXT, not JSON!
// RustDesk clients expect Content-Type: text/plain with body "SYSINFO_UPDATED" or "ID_NOT_FOUND".
func Sysinfo(c *gin.Context) {
	var req SysinfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Data(http.StatusBadRequest, "text/plain; charset=utf-8", []byte("INVALID_INPUT"))
		return
	}

	if req.ID == "" {
		c.Data(http.StatusBadRequest, "text/plain; charset=utf-8", []byte("INVALID_INPUT"))
		return
	}

	// Find peer by id and uuid
	peer, err := service.FindPeerByPeerIDAndUUID(req.ID, req.UUID)
	if err != nil {
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("ID_NOT_FOUND"))
		return
	}

	// Update system information
	if err := service.UpdatePeerSysinfo(peer, req.CPU, req.Memory, req.OS, req.Hostname, req.Username, req.Version); err != nil {
		logrus.WithError(err).Error("failed to update peer sysinfo")
		c.Data(http.StatusInternalServerError, "text/plain; charset=utf-8", []byte("SERVER_ERROR"))
		return
	}

	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("SYSINFO_UPDATED"))
}

// SysinfoVersion is a monotonically increasing version number for the sysinfo schema.
// Increment this when the sysinfo fields change to force clients to re-upload.
const SysinfoVersion = 1

// SysinfoVer handles GET /api/sysinfo_ver
// Returns the current sysinfo version number. Clients compare this to decide
// whether to re-upload their system information.
func SysinfoVer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ver": SysinfoVersion,
	})
}
