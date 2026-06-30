package service

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
)

const (
	// TOTP issuer label displayed in authenticator apps.
	TFADefaultIssuer = "RustDesk"
	// TfaSessionExpireMinutes is the lifetime of a TFA session token in minutes.
	TfaSessionExpireMinutes = 15
	// TfaSetupExpireMinutes is the lifetime of a TFA setup record in minutes.
	TfaSetupExpireMinutes = 5
)

// GenerateTOTPSecret creates a new TOTP key for the given user.
// Returns the base32-encoded secret and the otpauth:// URL for QR code generation.
func GenerateTOTPSecret(accountName string) (secret string, qrURL string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      TFADefaultIssuer,
		AccountName: accountName,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	return key.Secret(), key.URL(), nil
}

// ValidateTOTPCode checks whether the given code is valid for the secret.
// Uses the default 30-second window with one step of drift tolerance on each side.
func ValidateTOTPCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

// CreateTfaSetupToken stores a temporary TOTP secret for the enable flow.
// If a setup token already exists for this user, it is overwritten.
func CreateTfaSetupToken(userID uint, secret, qrURL string) error {
	expiresAt := time.Now().Add(time.Duration(TfaSetupExpireMinutes) * time.Minute).Unix()

	// Upsert: delete any existing setup for this user, then create new one
	database.DB.Where("user_id = ?", userID).Delete(&model.TfaSetupToken{})

	record := &model.TfaSetupToken{
		UserID:    userID,
		Secret:    secret,
		QrURL:     qrURL,
		ExpiresAt: expiresAt,
	}

	return database.DB.Create(record).Error
}

// GetTfaSetupToken looks up the active TFA setup record for a user.
// Returns nil if not found or expired.
func GetTfaSetupToken(userID uint) (*model.TfaSetupToken, error) {
	var record model.TfaSetupToken
	if err := database.DB.Where("user_id = ? AND expires_at > ?", userID, time.Now().Unix()).
		First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// DeleteTfaSetupToken removes the TFA setup record for a user.
func DeleteTfaSetupToken(userID uint) error {
	return database.DB.Where("user_id = ?", userID).Delete(&model.TfaSetupToken{}).Error
}

// EnableTFAForUser persists the TOTP secret on the user record.
func EnableTFAForUser(userID uint, secret string) error {
	return database.DB.Model(&model.User{}).Where("id = ?", userID).
		Update("tfa_secret", secret).Error
}

// DisableTFAForUser clears the TOTP secret on the user record.
func DisableTFAForUser(userID uint) error {
	return database.DB.Model(&model.User{}).Where("id = ?", userID).
		Update("tfa_secret", "").Error
}

// CreateTfaSessionToken creates a temporary session token for the TOTP login flow.
// This is a short-lived UserToken used to bridge password auth and TOTP verification.
// Returns the token string (session_token) or an error.
func CreateTfaSessionToken(userID uint) (string, error) {
	tokenStr, err := GenerateToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(time.Duration(TfaSessionExpireMinutes) * time.Minute).Unix()

	sessionToken := &model.UserToken{
		UserID:     userID,
		DeviceUUID: "tfa-session",
		Token:      tokenStr,
		ExpiredAt:  expiresAt,
	}

	if err := database.DB.Create(sessionToken).Error; err != nil {
		return "", fmt.Errorf("failed to create TFA session token: %w", err)
	}

	return tokenStr, nil
}

// ValidateTfaSessionToken looks up a TFA session token and returns the associated user.
// Only accepts tokens with DeviceUUID == "tfa-session".
// Returns the user on success, or an error.
func ValidateTfaSessionToken(tokenStr string) (*model.User, error) {
	var token model.UserToken
	if err := database.DB.Where("token = ? AND device_uuid = ?", tokenStr, "tfa-session").
		First(&token).Error; err != nil {
		return nil, fmt.Errorf("session token not found")
	}

	now := time.Now().Unix()
	if now > token.ExpiredAt {
		database.DB.Delete(&token)
		return nil, fmt.Errorf("session token expired")
	}

	var user model.User
	if err := database.DB.Where("id = ?", token.UserID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return &user, nil
}

// DeleteTfaSessionToken removes a TFA session token.
func DeleteTfaSessionToken(tokenStr string) error {
	return database.DB.Where("token = ? AND device_uuid = ?", tokenStr, "tfa-session").
		Delete(&model.UserToken{}).Error
}

// GenerateRandomCode generates a cryptographically random 6-digit code.
func GenerateRandomCode() (string, error) {
	b := make([]byte, 4)
	if _, err := cryptorand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	code := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	return fmt.Sprintf("%06d", code%1000000), nil
}

// GenerateResetToken creates a cryptographically random 32-byte hex token.
func GenerateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := cryptorand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate reset token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
