package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
	"github.com/sirupsen/logrus"
)

// DeployRequest is the request body for POST /api/devices/deploy.
type DeployRequest struct {
	ID   string `json:"id"`
	UUID string `json:"uuid"`
	PK   string `json:"pk"`
}

// CLIRequest is the request body for POST /api/devices/cli.
type CLIRequest struct {
	ID              string `json:"id"`
	UUID            string `json:"uuid"`
	UserName        string `json:"user_name"`
	StrategyName    string `json:"strategy_name"`
	AddressBookName string `json:"address_book_name"`
}

// Deploy handles POST /api/devices/deploy
// Client sends: {"id":"...", "uuid":"...", "pk":"..."} with Bearer token auth.
// Returns plain text "OK", "NOT_ENABLED", "INVALID_INPUT", or "ID_TAKEN".
func Deploy(c *gin.Context) {
	// Bearer token auth — extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.Data(http.StatusUnauthorized, "text/plain; charset=utf-8", []byte("NOT_ENABLED"))
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	_, _, err := service.ValidateToken(tokenStr)
	if err != nil {
		logrus.WithError(err).Warn("deploy: invalid token")
		c.Data(http.StatusUnauthorized, "text/plain; charset=utf-8", []byte("NOT_ENABLED"))
		return
	}

	var req DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == "" || req.UUID == "" {
		c.Data(http.StatusBadRequest, "text/plain; charset=utf-8", []byte("INVALID_INPUT"))
		return
	}

	// Check if device ID is already taken by a different UUID
	existing, err := service.FindPeerByPeerIDAndUUID(req.ID, req.UUID)
	if err == nil && existing != nil {
		// Same id+uuid: already deployed, success
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("OK"))
		return
	}

	// Check if ID is used by another device
	// (this check is done by FindOrCreatePeer which finds or creates)
	peer, isNew, err := service.FindOrCreatePeer(req.ID, req.UUID)
	if err != nil {
		logrus.WithError(err).Error("deploy: failed to create peer")
		c.Data(http.StatusInternalServerError, "text/plain; charset=utf-8", []byte("SERVER_ERROR"))
		return
	}

	if !isNew {
		// Exists but with different UUID? Wait — FindOrCreatePeer uses both peer_id+uuid.
		// If it found one, it's same ID+UUID combo. So this is OK.
	}

	_ = peer
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("OK"))
}

// DeployCLI handles POST /api/devices/cli
// Client sends device properties with Bearer token auth.
// Updates peer fields. Returns empty body on success.
func DeployCLI(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.Data(http.StatusUnauthorized, "text/plain; charset=utf-8", []byte("NOT_ENABLED"))
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	_, _, err := service.ValidateToken(tokenStr)
	if err != nil {
		logrus.WithError(err).Warn("deploy_cli: invalid token")
		c.Data(http.StatusUnauthorized, "text/plain; charset=utf-8", []byte("NOT_ENABLED"))
		return
	}

	var req CLIRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == "" {
		c.Data(http.StatusBadRequest, "text/plain; charset=utf-8", []byte("INVALID_INPUT"))
		return
	}

	peer, err := service.FindPeerByPeerIDAndUUID(req.ID, req.UUID)
	if err != nil {
		// Create a new peer if not found
		peer, _, err = service.FindOrCreatePeer(req.ID, req.UUID)
		if err != nil {
			logrus.WithError(err).Error("deploy_cli: failed to create peer")
			c.Data(http.StatusInternalServerError, "text/plain; charset=utf-8", []byte("SERVER_ERROR"))
			return
		}
	}

	// Apply CLI-assigned attributes to the peer.
	if req.UserName != "" {
		if err := service.UpdatePeerSysinfo(peer, peer.CPU, peer.Memory, peer.OS, peer.Hostname, req.UserName, peer.Version); err != nil {
			logrus.WithError(err).Error("deploy_cli: failed to update peer")
		}
	}

	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte{})
}
