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
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
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
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Username and password are required",
		})
		return
	}

	user, err := service.LoginByPassword(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": err.Error(),
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
		IP:       c.ClientIP(),
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
// Placeholder — returns captcha not required for now. Full captcha in P1.
func Captcha(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"captcha_required": false,
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
