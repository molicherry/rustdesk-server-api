package server

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/config"
	"github.com/rustdesk/rustdesk-api-server/internal/middleware"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
	"github.com/rustdesk/rustdesk-api-server/internal/wsutil"
	"github.com/sirupsen/logrus"
)

// NewServer creates and configures a new Gin engine with middleware.
func NewServer(cfg *config.Config) *gin.Engine {
	if cfg.Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	// Register routes (this also creates the WebSocket Hub)
	RegisterRoutes(r, cfg)

	// Start online status tracker in background
	go runOnlineStatusTracker()

	return r
}

// runOnlineStatusTracker periodically checks for peers that have stopped
// sending heartbeats and marks them as offline. It broadcasts status changes
// via WebSocket to all authenticated admin clients.
func runOnlineStatusTracker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	logrus.Info("online status tracker started")
	for range ticker.C {
		offlineIDs := service.CheckOfflinePeers()

		hub := wsutil.GlobalHub
		if hub == nil {
			continue
		}

		for _, peerID := range offlineIDs {
			hub.BroadcastToAuthenticated("peer_offline", gin.H{
				"id":   peerID,
				"time": 0,
			})
		}

		if len(offlineIDs) > 0 {
			logrus.WithField("count", len(offlineIDs)).Debug("peers marked offline")
		}
	}
}
