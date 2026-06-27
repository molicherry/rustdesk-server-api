package jwt

import (
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// CustomClaims contains the JWT claims.
type CustomClaims struct {
	UserID uint `json:"user_id"`
	jwtlib.RegisteredClaims
}

// GenerateToken creates a signed JWT token for the given user ID.
func GenerateToken(userID uint, key string, expireHours int) (string, error) {
	if key == "" {
		return "", fmt.Errorf("jwt key must not be empty")
	}

	claims := CustomClaims{
		UserID: userID,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(key))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return signed, nil
}

// ParseToken validates and parses a JWT token string.
func ParseToken(tokenStr, key string) (*CustomClaims, error) {
	token, err := jwtlib.ParseWithClaims(tokenStr, &CustomClaims{}, func(t *jwtlib.Token) (any, error) {
		if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(key), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	return claims, nil
}
