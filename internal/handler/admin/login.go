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
	Token         string       `json:"token"`
	User          UserResponse `json:"user"`
	Organizations []OrgInfo    `json:"organizations"`
}

// OrgInfo is a lightweight organization summary returned in login/user responses.
type OrgInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// UserResponse is the public user representation returned in API responses.
type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	Role     string `json:"role"`
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
		Role:     u.Role,
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

	// TOTP check: if the user has TFA enabled, require a TOTP code before issuing a token.
	if user.TFASecret != "" {
		sessionToken, err := service.CreateTfaSessionToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "server_error",
				"message": "Failed to create session token",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"totp_required": true,
			"session_token": sessionToken,
		})
		return
	}

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

	// Build organizations list for the response
	orgs := getOrgInfoList(user.ID)

	c.JSON(http.StatusOK, LoginResponse{
		Token:         token.Token,
		User:          userResponse(user),
		Organizations: orgs,
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

// VerifyTotpRequest is the request body for TOTP login verification.
type VerifyTotpRequest struct {
	SessionToken string `json:"session_token" binding:"required"`
	Code         string `json:"code" binding:"required"`
}

// VerifyTotpLogin handles POST /api/admin/login/verify-totp
// Validates the session token and TOTP code, then issues a real API token.
func VerifyTotpLogin(c *gin.Context) {
	ip := c.ClientIP()

	var req VerifyTotpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "session_token and code are required",
		})
		return
	}

	// Validate the session token (TFA session tokens only)
	user, err := service.ValidateTfaSessionToken(req.SessionToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid or expired session token",
		})
		return
	}

	// Validate the TOTP code
	if !service.ValidateTOTPCode(user.TFASecret, req.Code) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid_totp_code",
			"message": "Invalid TOTP code",
		})
		return
	}

	// TOTP verified — delete the session token and create a real API token
	_ = service.DeleteTfaSessionToken(req.SessionToken)

	token, err := service.CreateToken(user.ID, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to create token",
		})
		return
	}

	// Record login log with TOTP type
	loginLog := &model.LoginLog{
		UserID:   user.ID,
		Client:   c.Request.UserAgent(),
		IP:       ip,
		Type:     "totp",
		Platform: "web",
	}
	database.DB.Create(loginLog)

	// Build organizations list for the response
	orgs := getOrgInfoList(user.ID)

	c.JSON(http.StatusOK, LoginResponse{
		Token:         token.Token,
		User:          userResponse(user),
		Organizations: orgs,
	})
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
	orgs := getOrgInfoList(u.ID)

	c.JSON(http.StatusOK, gin.H{
		"user":          userResponse(u),
		"organizations": orgs,
	})
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
		_ = service.DeleteUserTokens(u.ID)
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

// getOrgInfoList builds the list of organizations a user belongs to.
func getOrgInfoList(userID uint) []OrgInfo {
	memberships, err := service.ListUserOrganizations(userID)
	if err != nil || len(memberships) == 0 {
		return []OrgInfo{}
	}

	orgs := make([]OrgInfo, 0, len(memberships))
	for _, m := range memberships {
		org, err := service.FindOrganizationByID(m.OrganizationID)
		if err != nil {
			continue
		}
		orgs = append(orgs, OrgInfo{
			ID:   org.ID,
			Name: org.Name,
			Role: m.Role,
		})
	}
	return orgs
}
