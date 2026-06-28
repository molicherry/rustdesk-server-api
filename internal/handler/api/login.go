package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/middleware"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
)

// ClientLoginRequest is the request body from the RustDesk client for /api/login.
type ClientLoginRequest struct {
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	ID           string `json:"id"`
	UUID         string `json:"uuid"`
	CaptchaToken string `json:"captcha_token,omitempty"`
}

// ClientLoginResponse is the response for a successful client login.
// MUST match the RustDesk client protocol exactly.
type ClientLoginResponse struct {
	AccessToken string `json:"access_token"`
	Type        string `json:"type"`
}

// ClientLogin handles POST /api/login
// RustDesk clients send {username, password, id, uuid} and expect
// {access_token: "...", type: "access_token"} in response.
func ClientLogin(c *gin.Context) {
	ip := c.ClientIP()

	var req ClientLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid request",
		})
		return
	}

	// Rate limiting: block if too many failed attempts from this IP.
	if middleware.LoginLimiter.IsBlocked(ip) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":   "too_many_requests",
			"message": "Too many login attempts. Please try again later.",
		})
		return
	}

	user, err := service.LoginByPassword(req.Username, req.Password)
	if err != nil {
		// Record failed attempt.
		middleware.LoginLimiter.RecordAttempt(ip)

		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid username or password",
		})
		return
	}

	// Successful login — clear rate limiter for this IP.
	middleware.LoginLimiter.Clear(ip)

	// Create a token for the client with device UUID
	token, err := service.CreateToken(user.ID, req.UUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to create token",
		})
		return
	}

	// Record login log
	loginLog := &model.LoginLog{
		UserID:   user.ID,
		DeviceID: req.ID,
		UUID:     req.UUID,
		Client:   c.Request.UserAgent(),
		IP:       ip,
		Type:     "password",
		Platform: "rustdesk",
	}
	database.DB.Create(loginLog)

	c.JSON(http.StatusOK, ClientLoginResponse{
		AccessToken: token.Token,
		Type:        "access_token",
	})
}

// ClientLogout handles POST /api/logout
func ClientLogout(c *gin.Context) {
	// For now, just accept — full RustAuth will be in Phase 2
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

// LoginOptions handles HEAD and GET /api/login-options
// RustDesk clients send HEAD for TLS warmup, GET for available login methods.
// Returns an empty object for now; OIDC providers in P1.
func LoginOptions(c *gin.Context) {
	if c.Request.Method == http.MethodHead {
		c.Status(http.StatusOK)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
