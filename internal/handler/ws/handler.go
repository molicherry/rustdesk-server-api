package ws

import (
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
func WSUpgradeHandler(hub *wsutil.Hub) gin.HandlerFunc {
	wsutil.GlobalHub = hub

	return func(c *gin.Context) {
		conn, err := wsutil.Upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logrus.WithError(err).Warn("websocket upgrade failed")
			return
		}

		client := &wsutil.Client{
			Hub:  hub,
			Conn: conn,
			Send: make(chan []byte, 256),
		}

		hub.Register(client)

		go client.WritePump()
		go client.ReadPump()
	}
}
