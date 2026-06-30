package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
	"slices"
)

// ContextKeyOrgID is the gin context key for the current organization ID.
const ContextKeyOrgID = "currentOrgID"

// RequireRole returns middleware that checks the authenticated user's system-level role.
// Must be applied after BackendUserAuth.
func RequireRole(roles ...string) gin.HandlerFunc {
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
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Invalid user context",
			})
			return
		}

		if !slices.Contains(roles, u.Role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Insufficient role",
			})
			return
		}

		c.Next()
	}
}

// RequireOrgRole returns middleware that checks the user's role within a specific organization.
// The organization ID is extracted from the query parameter "org_id".
// When org_id=0 or missing, only admin role can proceed (global view).
// Must be applied after BackendUserAuth.
func RequireOrgRole(roles ...string) gin.HandlerFunc {
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
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Invalid user context",
			})
			return
		}

		orgID := getOrgIDFromRequest(c)

		// orgID=0 means global view — only system-level admin can access
		if orgID == 0 {
			if u.Role != "admin" {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error":   "forbidden",
					"message": "Global view requires admin role",
				})
				return
			}
			c.Set(ContextKeyOrgID, uint(0))
			c.Next()
			return
		}

		// Check user's role within the organization
		uo, err := service.FindUserOrganization(u.ID, orgID)
		if err != nil || !slices.Contains(roles, uo.Role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Insufficient organization role",
			})
			return
		}

		c.Set(ContextKeyOrgID, orgID)
		c.Next()
	}
}

// getOrgIDFromRequest extracts the organization ID from the request.
// Looks in: query param "org_id", then path param "orgID".
func getOrgIDFromRequest(c *gin.Context) uint {
	// First check query param
	if orgIDStr := c.Query("org_id"); orgIDStr != "" {
		if id, err := strconv.ParseUint(orgIDStr, 10, 64); err == nil {
			return uint(id)
		}
	}
	// Then check path param
	if orgIDStr := c.Param("orgID"); orgIDStr != "" {
		if id, err := strconv.ParseUint(orgIDStr, 10, 64); err == nil {
			return uint(id)
		}
	}
	return 0
}
