package wsutil

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096
)

// Client represents a single WebSocket connection.
type Client struct {
	Hub           *Hub
	Conn          *websocket.Conn
	Send          chan []byte
	Authenticated bool
	UserID        uint
}

// ReadPump pumps messages from the WebSocket connection to the hub.
// The first message must be an auth message: {"type":"auth","token":"..."}
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logrus.WithError(err).Warn("websocket read error")
			}
			break
		}

		// First message must be auth
		if !c.Authenticated {
			c.handleAuth(message)
			continue
		}

		// After auth, messages are heartbeats — just reset deadline
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	}
}

// handleAuth validates the auth message and marks the client as authenticated.
func (c *Client) handleAuth(message []byte) {
	var msg struct {
		Type  string `json:"type"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(message, &msg); err != nil || msg.Type != "auth" {
		c.Send <- []byte(`{"type":"error","message":"invalid auth message"}`)
		return
	}

	// Validate token via service layer — imported in handler/ws to avoid circular deps
	// The validation callback is set by the ws handler
	if HubValidateToken != nil {
		userID, ok := HubValidateToken(msg.Token)
		if ok {
			c.Authenticated = true
			c.UserID = userID
			c.Send <- []byte(`{"type":"auth_ok"}`)
			logrus.WithField("user_id", userID).Info("websocket client authenticated")
			return
		}
	}

	c.Send <- []byte(`{"type":"error","message":"authentication failed"}`)
}

// WritePump pumps messages from the hub to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logrus.WithError(err).Warn("websocket write error")
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
}

// GlobalHub is the singleton WebSocket hub instance, set by server initialization.
var GlobalHub *Hub

// HubValidateToken is a callback set by the ws handler to validate WebSocket auth tokens.
// Returns (userID, true) on success, (0, false) on failure.
var HubValidateToken func(token string) (uint, bool)

// NewHub creates a new Hub instance and starts its event loop in a background goroutine.
func NewHub() *Hub {
	h := &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
	}
	return h
}

// Run starts the hub event loop. Must be called in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			logrus.WithField("total", len(h.clients)).Info("websocket client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			logrus.WithField("total", len(h.clients)).Info("websocket client disconnected")

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if client.Authenticated {
					select {
					case client.Send <- message:
					default:
						// Client buffer full — drop them
						go func(c *Client) {
							h.unregister <- c
						}(client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(c *Client) {
	h.register <- c
}

// BroadcastToAuthenticated sends a message to all authenticated clients.
func (h *Hub) BroadcastToAuthenticated(eventType string, data any) {
	msg, err := json.Marshal(map[string]any{
		"type": eventType,
		"data": data,
	})
	if err != nil {
		logrus.WithError(err).Error("failed to marshal broadcast message")
		return
	}

	select {
	case h.broadcast <- msg:
	default:
		logrus.Warn("broadcast channel full, dropping message")
	}
}

// ClientCount returns the number of currently connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Upgrader is the WebSocket connection upgrader with default settings.
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for development; restrict in production via config
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
