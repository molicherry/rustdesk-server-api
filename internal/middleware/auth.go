package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
)

const (
	// ContextKeyUser is the gin context key for the current authenticated user.
	ContextKeyUser = "currentUser"
	// ContextKeyToken is the gin context key for the current API token.
	ContextKeyToken = "currentToken"
)

// BackendUserAuth is a middleware that authenticates requests using the api-token header.
// It looks up the token in the user_tokens table, validates expiry, and injects the
// user and token into the gin context.
func BackendUserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("api-token")
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Missing api-token header",
			})
			return
		}

		user, token, err := service.ValidateToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": err.Error(),
			})
			return
		}

		c.Set(ContextKeyUser, user)
		c.Set(ContextKeyToken, token)
		c.Next()
	}
}
