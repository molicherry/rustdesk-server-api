package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/internal/middleware"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
)

// TfaEnableRequest is the request body for enabling TFA.
type TfaEnableRequest struct {
	// No fields needed — TFA is enabled for the current user.
}

// TfaVerifyRequest is the request body for verifying TFA setup.
type TfaVerifyRequest struct {
	Code string `json:"code" binding:"required"`
}

// TfaDisableRequest is the request body for disabling TFA.
type TfaDisableRequest struct {
	Code string `json:"code" binding:"required"`
}

// EnableTfa handles POST /api/admin/user/tfa/enable
// Generates a TOTP secret and QR code URL for the current user.
func EnableTfa(c *gin.Context) {
	user, exists := c.Get(middleware.ContextKeyUser)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	u := user.(*model.User)

	// Check if TFA is already enabled
	if u.TFASecret != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "TFA is already enabled. Disable it first to re-enable.",
		})
		return
	}

	secret, qrURL, err := service.GenerateTOTPSecret(u.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to generate TOTP secret",
		})
		return
	}

	if err := service.CreateTfaSetupToken(u.ID, secret, qrURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to store TFA setup",
		})
		return
	}

	// Return secret and QR URL for frontend to display
	c.JSON(http.StatusOK, gin.H{
		"secret":      secret,
		"qr_code_url": qrURL,
	})
}

// VerifyTfa handles POST /api/admin/user/tfa/verify
// Validates the TOTP code against the temporarily stored secret and binds TFA to the user.
func VerifyTfa(c *gin.Context) {
	user, exists := c.Get(middleware.ContextKeyUser)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	u := user.(*model.User)

	var req TfaVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Code is required",
		})
		return
	}

	// Look up the pending setup token
	setupToken, err := service.GetTfaSetupToken(u.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "No pending TFA setup found. Please enable TFA first.",
		})
		return
	}

	// Validate the TOTP code against the pending secret
	if !service.ValidateTOTPCode(setupToken.Secret, req.Code) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_totp_code",
			"message": "Invalid TOTP code. Please try again.",
		})
		return
	}

	// Code is valid — enable TFA on the user record
	if err := service.EnableTFAForUser(u.ID, setupToken.Secret); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to enable TFA",
		})
		return
	}

	// Clean up the setup token
	_ = service.DeleteTfaSetupToken(u.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "TFA enabled successfully",
	})
}

// DisableTfa handles POST /api/admin/user/tfa/disable
// Disables TFA for the current user after verifying the current TOTP code.
func DisableTfa(c *gin.Context) {
	user, exists := c.Get(middleware.ContextKeyUser)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	u := user.(*model.User)

	if u.TFASecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "TFA is not enabled",
		})
		return
	}

	var req TfaDisableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Code is required",
		})
		return
	}

	// Verify the TOTP code before disabling
	if !service.ValidateTOTPCode(u.TFASecret, req.Code) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_totp_code",
			"message": "Invalid TOTP code",
		})
		return
	}

	if err := service.DisableTFAForUser(u.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to disable TFA",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "TFA disabled successfully",
	})
}
