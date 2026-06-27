package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
)

// AdminPrivilege is a middleware that checks if the authenticated user has admin privileges.
// It must be applied after BackendUserAuth.
func AdminPrivilege() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get(ContextKeyUser)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Authentication required",
			})
			return
		}

		u, ok := user.(*model.User)
		if !ok || !u.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Admin privileges required",
			})
			return
		}

		c.Next()
	}
}
