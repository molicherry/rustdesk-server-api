package service

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	mathrand "math/rand/v2"
	"time"

	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"golang.org/x/crypto/bcrypt"
)

const (
	// TokenBytes is the number of random bytes for API tokens (32 bytes → 64 hex chars).
	TokenBytes = 32
	// TokenExpireHours is the default token expiry in hours (7 days).
	TokenExpireHours = 168
)

// HashPassword creates a bcrypt hash of the given plaintext password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// VerifyPassword checks whether a plaintext password matches a bcrypt hash.
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken creates a cryptographically random hex-encoded token.
func GenerateToken() (string, error) {
	b := make([]byte, TokenBytes)
	if _, err := cryptorand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// GenerateRandomPassword creates a random 8-character alphanumeric password.
func GenerateRandomPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[mathrand.IntN(len(charset))]
	}
	return string(b)
}

// LoginByPassword looks up a user by username and verifies the password.
// Returns the user on success, or an error.
func LoginByPassword(username, password string) (*model.User, error) {
	var user model.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	if user.Status != 1 {
		return nil, fmt.Errorf("invalid username or password")
	}

	if !VerifyPassword(password, user.Password) {
		return nil, fmt.Errorf("invalid username or password")
	}

	return &user, nil
}

// CreateToken generates a new API token for the given user and saves it to the database.
func CreateToken(userID uint, deviceUUID string) (*model.UserToken, error) {
	tokenStr, err := GenerateToken()
	if err != nil {
		return nil, err
	}

	token := &model.UserToken{
		UserID:     userID,
		DeviceUUID: deviceUUID,
		Token:      tokenStr,
		ExpiredAt:  time.Now().Add(time.Duration(TokenExpireHours) * time.Hour).Unix(),
	}

	if err := database.DB.Create(token).Error; err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	return token, nil
}

// ValidateToken looks up a token string, checks expiry, and auto-refreshes if within 1/3 of remaining time.
// Returns the associated user and token on success.
func ValidateToken(tokenStr string) (*model.User, *model.UserToken, error) {
	var token model.UserToken
	if err := database.DB.Where("token = ?", tokenStr).First(&token).Error; err != nil {
		return nil, nil, fmt.Errorf("token not found")
	}

	now := time.Now().Unix()
	if now > token.ExpiredAt {
		database.DB.Delete(&token)
		return nil, nil, fmt.Errorf("token expired")
	}

	var user model.User
	if err := database.DB.Where("id = ?", token.UserID).First(&user).Error; err != nil {
		return nil, nil, fmt.Errorf("user not found")
	}

	if user.Status != 1 {
		return nil, nil, fmt.Errorf("account is disabled")
	}

	// Auto-refresh: if within 1/3 of remaining lifetime, extend expiry
	totalLife := int64(TokenExpireHours * 3600)
	remaining := token.ExpiredAt - now
	if remaining < totalLife/3 {
		token.ExpiredAt = time.Now().Add(time.Duration(TokenExpireHours) * time.Hour).Unix()
		database.DB.Model(&token).Update("expired_at", token.ExpiredAt)
	}

	return &user, &token, nil
}

// DeleteToken removes a token from the database.
func DeleteToken(tokenStr string) error {
	result := database.DB.Where("token = ?", tokenStr).Delete(&model.UserToken{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete token: %w", result.Error)
	}
	return nil
}

// DeleteUserTokens removes all tokens for a given user (e.g., after password change).
func DeleteUserTokens(userID uint) error {
	result := database.DB.Where("user_id = ?", userID).Delete(&model.UserToken{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete user tokens: %w", result.Error)
	}
	return nil
}

// CreateUser creates a new user with a bcrypt-hashed password.
func CreateUser(username, password string, isAdmin bool) (*model.User, error) {
	// Check if username already exists
	var existing model.User
	if err := database.DB.Where("username = ?", username).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("username %q already exists", username)
	}

	hashed, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: username,
		Password: hashed,
		IsAdmin:  isAdmin,
		Status:   1,
	}

	if err := database.DB.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// FindUserByID looks up a user by primary key.
func FindUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}
