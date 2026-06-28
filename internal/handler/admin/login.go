package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/middleware"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
)

// LoginRequest is the request body for admin login.
type LoginRequest struct {
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	CaptchaToken string `json:"captcha_token,omitempty"`
}

// LoginResponse is the response body for a successful admin login.
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse is the public user representation returned in API responses.
type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Status   int    `json:"status"`
}

// userResponse converts a model.User to the API response struct.
func userResponse(u *model.User) UserResponse {
	return UserResponse{
		ID:       u.ID,
		Username: u.Username,
		IsAdmin:  u.IsAdmin,
		Email:    u.Email,
		Nickname: u.Nickname,
		Status:   u.Status,
	}
}

// AdminLogin handles POST /api/admin/login
func AdminLogin(c *gin.Context) {
	ip := c.ClientIP()

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Username and password are required",
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

	// Captcha check: if threshold is reached, captcha token is mandatory.
	if middleware.LoginLimiter.CaptchaRequired(ip) && req.CaptchaToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "captcha_required",
			"message": "Captcha token is required. Solve the captcha and include captcha_token in the request.",
		})
		return
	}
	// TODO: validate captcha token against a captcha service when implemented.
	// For now, any non-empty captcha_token passes when captcha is required.

	user, err := service.LoginByPassword(req.Username, req.Password)
	if err != nil {
		// Record failed attempt.
		middleware.LoginLimiter.RecordAttempt(ip)

		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": err.Error(),
		})
		return
	}

	// Successful login — clear rate limiter for this IP.
	middleware.LoginLimiter.Clear(ip)

	token, err := service.CreateToken(user.ID, "")
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
		Client:   c.Request.UserAgent(),
		IP:       ip,
		Type:     "password",
		Platform: "web",
	}
	database.DB.Create(loginLog)

	c.JSON(http.StatusOK, LoginResponse{
		Token: token.Token,
		User:  userResponse(user),
	})
}

// AdminLogout handles POST /api/admin/logout
func AdminLogout(c *gin.Context) {
	tokenStr := c.GetHeader("api-token")
	if tokenStr == "" {
		c.JSON(http.StatusOK, gin.H{"message": "Already logged out"})
		return
	}

	if err := service.DeleteToken(tokenStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to logout",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Captcha handles GET /api/admin/captcha
// Returns whether captcha is required based on the rate limiter threshold.
func Captcha(c *gin.Context) {
	ip := c.ClientIP()
	required := middleware.LoginLimiter.CaptchaRequired(ip)

	c.JSON(http.StatusOK, gin.H{
		"captcha_required": required,
	})
}

// UserRegister handles POST /api/admin/user/register
// Placeholder — will be implemented when registration is enabled.
func UserRegister(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "User registration is not yet implemented",
	})
}

// GetCurrentUser handles GET /api/admin/user/current
func GetCurrentUser(c *gin.Context) {
	user, exists := c.Get(middleware.ContextKeyUser)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	u := user.(*model.User)
	c.JSON(http.StatusOK, userResponse(u))
}

// ChangeCurrentPasswordRequest is the request body for changing password.
type ChangeCurrentPasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ChangeCurrentPassword handles POST /api/admin/user/changeCurPwd
func ChangeCurrentPassword(c *gin.Context) {
	user, exists := c.Get(middleware.ContextKeyUser)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	token, _ := c.Get(middleware.ContextKeyToken)

	var req ChangeCurrentPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "old_password and new_password are required",
		})
		return
	}

	u := user.(*model.User)

	// Verify old password
	if !service.VerifyPassword(req.OldPassword, u.Password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Old password is incorrect",
		})
		return
	}

	// Hash new password
	hashed, err := service.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to hash password",
		})
		return
	}

	// Update password
	if err := database.DB.Model(u).Update("password", hashed).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to update password",
		})
		return
	}

	// Invalidate all other tokens (keep current one)
	if t, ok := token.(*model.UserToken); ok {
		database.DB.Where("user_id = ? AND id != ?", u.ID, t.ID).Delete(&model.UserToken{})
	} else {
		service.DeleteUserTokens(u.ID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// ConfigServer handles GET /api/admin/config/server
func ConfigServer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id_server":    "",
		"relay_server": "",
		"api_server":   "",
	})
}
