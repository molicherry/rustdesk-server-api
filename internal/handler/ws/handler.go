package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
	"github.com/rustdesk/rustdesk-api-server/internal/wsutil"
	"github.com/sirupsen/logrus"
)

func init() {
	// Register the token validation callback for WebSocket hub auth.
	wsutil.HubValidateToken = func(tokenStr string) (uint, bool) {
		user, _, err := service.ValidateToken(tokenStr)
		if err != nil {
			return 0, false
		}
		return user.ID, true
	}
}

// WSUpgradeHandler returns an HTTP handler that upgrades requests to WebSocket
// connections and registers clients with the given Hub.
// Authentication is enforced BEFORE the WebSocket upgrade, rejecting
// unauthenticated requests with an HTTP 401 response.
func WSUpgradeHandler(hub *wsutil.Hub) gin.HandlerFunc {
	wsutil.GlobalHub = hub

	return func(c *gin.Context) {
		// Pre-upgrade authentication: validate the token BEFORE upgrading to
		// WebSocket, so unauthenticated connections never reach the hub.
		tokenStr := c.GetHeader("api-token")
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Missing api-token header or token query parameter",
			})
			return
		}

		user, _, err := service.ValidateToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": err.Error(),
			})
			return
		}

		// Store the authenticated user in the Gin context before upgrading.
		c.Set("currentUser", user)

		conn, err := wsutil.Upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logrus.WithError(err).Warn("websocket upgrade failed")
			return
		}

		// Client is already authenticated; set flags so ReadPump
		// skips the first-message auth path.
		client := &wsutil.Client{
			Hub:           hub,
			Conn:          conn,
			Send:          make(chan []byte, 256),
			Authenticated: true,
			UserID:        user.ID,
		}

		hub.Register(client)

		go client.WritePump()
		go client.ReadPump()
	}
}
