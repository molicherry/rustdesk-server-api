package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/config"
	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
)

// ForgotPasswordRequest is the request body for initiating password reset.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

// ResetPasswordRequest is the request body for completing password reset.
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required"`
	Token       string `json:"token" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ForgotPassword handles POST /api/admin/user/forgot-password (public)
// Sends a password reset code to the user's email.
func ForgotPassword(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ForgotPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "Email is required",
			})
			return
		}

		// Look up the user by email
		var user model.User
		if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
			// Don't reveal whether the email exists — always return success
			c.JSON(http.StatusOK, gin.H{
				"message": "If the email exists, a reset code has been sent",
			})
			return
		}

		if err := service.SendPasswordResetEmail(cfg.SMTP, req.Email); err != nil {
			if err.Error() == "smtp_not_configured" {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error":   "smtp_not_configured",
					"message": "Email service is not configured",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "server_error",
				"message": "Failed to send reset email",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "If the email exists, a reset code has been sent",
		})
	}
}

// ResetPassword handles POST /api/admin/user/reset-password (public)
// Resets the user's password after validating the reset token and code.
func ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Email, token, code, and new_password are required",
		})
		return
	}

	// Validate the reset token and code
	if !service.ValidatePasswordReset(req.Email, req.Token, req.Code) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_reset",
			"message": "Invalid or expired reset token/code",
		})
		return
	}

	// Find the user by email
	var user model.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "User not found",
		})
		return
	}

	// Hash and update the password
	hashed, err := service.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to hash password",
		})
		return
	}

	if err := database.DB.Model(&user).Update("password", hashed).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to update password",
		})
		return
	}

	// Invalidate all existing tokens for this user
	service.DeleteUserTokens(user.ID)

	// Clean up reset records
	service.ConsumePasswordReset(req.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "Password has been reset successfully",
	})
}
