package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/config"
	"github.com/rustdesk/rustdesk-api-server/internal/middleware"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
)

// SendVerificationRequest is the request body for sending a verification code.
type SendVerificationRequest struct {
	Email string `json:"email" binding:"required"`
}

// VerifyEmailRequest is the request body for verifying an email code.
type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// SendVerification handles POST /api/admin/user/send-verification
// Sends a 6-digit verification code to the specified email address.
func SendVerification(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get(middleware.ContextKeyUser)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Not authenticated",
			})
			return
		}

		var req SendVerificationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "Email is required",
			})
			return
		}

		if err := service.SendVerificationCode(cfg.SMTP, req.Email); err != nil {
			if err.Error() == "smtp_not_configured" {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error":   "smtp_not_configured",
					"message": "Email service is not configured",
				})
				return
			}
			if err.Error() == "too many requests: please wait before requesting another code" {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":   "too_many_requests",
					"message": "Please wait before requesting another code",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "server_error",
				"message": "Failed to send verification email",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Verification code sent",
		})
	}
}

// VerifyEmail handles POST /api/admin/user/verify-email
// Validates the verification code and marks the email as verified.
func VerifyEmail(c *gin.Context) {
	user, exists := c.Get(middleware.ContextKeyUser)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	u := user.(*model.User)

	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Email and code are required",
		})
		return
	}

	// Verify the code against the stored record
	if !service.VerifyEmailCode(req.Email, req.Code) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_code",
			"message": "Invalid or expired verification code",
		})
		return
	}

	// Mark email as verified on the user record
	if err := service.ConsumeEmailVerification(u.ID, req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to verify email",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}
