package service

import (
	"fmt"
	"log"
	"time"

	"github.com/go-gomail/gomail"
	"github.com/rustdesk/rustdesk-api-server/config"
	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
)

const (
	// EmailVerificationExpireMinutes is the lifetime of an email verification code in minutes.
	EmailVerificationExpireMinutes = 5
	// EmailThrottleSeconds enforces minimum interval between code sends to the same email.
	EmailThrottleSeconds = 60
	// PasswordResetExpireMinutes is the lifetime of a password reset token in minutes.
	PasswordResetExpireMinutes = 60
)

// IsSMTPConfigured checks whether SMTP is enabled and has required fields.
func IsSMTPConfigured(cfg config.SMTPConfig) bool {
	return cfg.Enable && cfg.Host != "" && cfg.Port > 0 && cfg.From != ""
}

// SendEmail sends a plain-text email via SMTP using gomail.
func SendEmail(cfg config.SMTPConfig, to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendVerificationCode generates a 6-digit code, persists it, and emails it to the user.
// Returns an error if the SMTP is not configured, throttled, or send fails.
func SendVerificationCode(cfg config.SMTPConfig, email string) error {
	if !IsSMTPConfigured(cfg) {
		return fmt.Errorf("smtp_not_configured")
	}

	// Throttle: ensure at least EmailThrottleSeconds since last code for this email
	var recent model.EmailVerification
	cutoff := time.Now().Add(-time.Duration(EmailThrottleSeconds) * time.Second).Unix()
	if err := database.DB.Where("email = ? AND created_at >= ?", email, cutoff).
		Order("created_at DESC").First(&recent).Error; err == nil {
		return fmt.Errorf("too many requests: please wait before requesting another code")
	}

	code, err := GenerateRandomCode()
	if err != nil {
		return err
	}

	// Delete old codes for this email
	database.DB.Where("email = ?", email).Delete(&model.EmailVerification{})

	expiresAt := time.Now().Add(time.Duration(EmailVerificationExpireMinutes) * time.Minute).Unix()

	record := &model.EmailVerification{
		Email:     email,
		Code:      code,
		ExpiresAt: expiresAt,
	}

	if err := database.DB.Create(record).Error; err != nil {
		return fmt.Errorf("failed to store verification code: %w", err)
	}

	subject := "RustDesk Email Verification Code"
	body := fmt.Sprintf("Your verification code is: %s\n\nThis code will expire in %d minutes.", code, EmailVerificationExpireMinutes)

	if err := SendEmail(cfg, email, subject, body); err != nil {
		// Log send failure but don't fail the whole operation — code is stored
		log.Printf("WARNING: failed to send verification email to %s: %v", email, err)
		return fmt.Errorf("failed to send verification email")
	}

	return nil
}

// VerifyEmailCode checks whether the given code is valid for the email address.
func VerifyEmailCode(email, code string) bool {
	var record model.EmailVerification
	if err := database.DB.Where("email = ? AND code = ? AND expires_at > ?",
		email, code, time.Now().Unix()).First(&record).Error; err != nil {
		return false
	}
	return true
}

// ConsumeEmailVerification marks the email as verified on the user record
// and cleans up the verification code.
func ConsumeEmailVerification(userID uint, email string) error {
	if err := database.DB.Model(&model.User{}).Where("id = ?", userID).
		Update("email_verified", true).Error; err != nil {
		return fmt.Errorf("failed to update email_verified: %w", err)
	}

	// Clean up verification records for this email
	database.DB.Where("email = ?", email).Delete(&model.EmailVerification{})

	return nil
}

// SendPasswordResetEmail generates a reset token and code, stores them, and sends the code via email.
func SendPasswordResetEmail(cfg config.SMTPConfig, email string) error {
	if !IsSMTPConfigured(cfg) {
		return fmt.Errorf("smtp_not_configured")
	}

	code, err := GenerateRandomCode()
	if err != nil {
		return err
	}

	resetToken, err := GenerateResetToken()
	if err != nil {
		return err
	}

	// Remove any existing reset tokens for this email
	database.DB.Where("email = ?", email).Delete(&model.PasswordReset{})

	expiresAt := time.Now().Add(time.Duration(PasswordResetExpireMinutes) * time.Minute).Unix()

	record := &model.PasswordReset{
		Email:     email,
		Token:     resetToken,
		Code:      code,
		ExpiresAt: expiresAt,
	}

	if err := database.DB.Create(record).Error; err != nil {
		return fmt.Errorf("failed to store password reset: %w", err)
	}

	subject := "RustDesk Password Reset"
	body := fmt.Sprintf(
		"Your password reset code is: %s\n\n"+
			"Use this code to reset your password. This code will expire in %d minutes.\n\n"+
			"If you did not request a password reset, please ignore this email.",
		code, PasswordResetExpireMinutes)

	if err := SendEmail(cfg, email, subject, body); err != nil {
		log.Printf("WARNING: failed to send password reset email to %s: %v", email, err)
		return fmt.Errorf("failed to send password reset email")
	}

	return nil
}

// ValidatePasswordReset checks token and code, returns the associated email on success.
func ValidatePasswordReset(email, token, code string) bool {
	var record model.PasswordReset
	if err := database.DB.Where("email = ? AND token = ? AND code = ? AND expires_at > ?",
		email, token, code, time.Now().Unix()).First(&record).Error; err != nil {
		return false
	}
	return true
}

// ConsumePasswordReset cleans up the reset record after a successful password change.
func ConsumePasswordReset(email string) {
	database.DB.Where("email = ?", email).Delete(&model.PasswordReset{})
}
