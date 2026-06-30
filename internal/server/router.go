package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/config"
	"github.com/rustdesk/rustdesk-api-server/internal/handler/admin"
	"github.com/rustdesk/rustdesk-api-server/internal/handler/api"
	wshandler "github.com/rustdesk/rustdesk-api-server/internal/handler/ws"
	"github.com/rustdesk/rustdesk-api-server/internal/middleware"
	"github.com/rustdesk/rustdesk-api-server/internal/wsutil"
)

// RegisterRoutes registers all API routes on the Gin engine.
func RegisterRoutes(r *gin.Engine, cfg *config.Config) {
	// Global middleware
	r.Use(middleware.Localization())

	// Health check
	r.GET("/api/version", versionHandler)
	r.GET("/api/health", healthHandler)

	// === Client API group (/api/*) ===
	apiGroup := r.Group("/api")
	apiGroup.Use(middleware.BackendUserAuth())
	{
		// login-options supports both HEAD (TLS warmup) and GET
		apiGroup.HEAD("/login-options", api.LoginOptions)
		apiGroup.GET("/login-options", api.LoginOptions)

		// Client login/logout
		apiGroup.POST("/login", api.ClientLogin)
		apiGroup.POST("/logout", api.ClientLogout)

		// Device heartbeat and sysinfo (no auth required)
		apiGroup.POST("/heartbeat", api.Heartbeat)
		apiGroup.POST("/sysinfo", api.Sysinfo)
		apiGroup.GET("/sysinfo_ver", api.SysinfoVer)

		// Device deployment (Bearer token auth handled in handler)
		apiGroup.POST("/devices/deploy", api.Deploy)
		apiGroup.POST("/devices/cli", api.DeployCLI)

		// Audit endpoints (Phase 3 — no auth, device-id based)
		apiGroup.POST("/audit/conn", api.AuditConn)
		apiGroup.POST("/audit/file", api.AuditFile)

		// OIDC endpoints (P1)
		apiGroup.POST("/oidc/auth", notImplemented)
		apiGroup.GET("/oidc/auth-query", notImplemented)

		// Client address book and peer endpoints (Phase 3 — device-id based)
		apiGroup.GET("/ab", api.ABGet)
		apiGroup.POST("/ab", api.ABSync)
		apiGroup.POST("/ab/personal", api.ABPersonal)
		apiGroup.POST("/ab/settings", api.ABSettings)
		apiGroup.POST("/ab/shared/profiles", api.ABSharedProfiles)
		apiGroup.POST("/ab/peers", api.ABPeers)
		apiGroup.POST("/ab/tags/:guid", api.ABTags)
		apiGroup.POST("/ab/peer/add/:guid", api.ABPeerAdd)
		apiGroup.DELETE("/ab/peer/:guid", api.ABPeerDelete)
		apiGroup.PUT("/ab/peer/update/:guid", api.ABPeerUpdate)
		apiGroup.POST("/ab/tag/add/:guid", api.ABTagAdd)
		apiGroup.PUT("/ab/tag/rename/:guid", api.ABTagRename)
		apiGroup.PUT("/ab/tag/update/:guid", api.ABTagUpdate)
		apiGroup.DELETE("/ab/tag/:guid", api.ABTagDelete)

		// RustAuth-protected client endpoints (Phase 2+)
		apiGroup.GET("/user/info", notImplemented)
		apiGroup.GET("/peers", notImplemented)
	}

	// === Admin API group (/api/admin/*) ===
	adminGroup := r.Group("/api/admin")
	{
		// Public endpoints (no auth)
		adminGroup.POST("/login", admin.AdminLogin)
		adminGroup.POST("/login/verify-totp", admin.VerifyTotpLogin)
		adminGroup.POST("/logout", admin.AdminLogout)
		adminGroup.GET("/captcha", admin.Captcha)
		adminGroup.POST("/user/register", admin.UserRegister)
		adminGroup.POST("/user/forgot-password", admin.ForgotPassword(cfg))
		adminGroup.POST("/user/reset-password", admin.ResetPassword)
		adminGroup.GET("/config/server", admin.ConfigServer)

		// Authenticated endpoints (BackendUserAuth required)
		authGroup := adminGroup.Group("")
		authGroup.Use(middleware.BackendUserAuth())
		{
			authGroup.GET("/user/current", admin.GetCurrentUser)
			authGroup.POST("/user/changeCurPwd", admin.ChangeCurrentPassword)

			// TOTP two-factor authentication
			authGroup.POST("/user/tfa/enable", admin.EnableTfa)
			authGroup.POST("/user/tfa/verify", admin.VerifyTfa)
			authGroup.POST("/user/tfa/disable", admin.DisableTfa)

			// Email verification
			authGroup.POST("/user/send-verification", admin.SendVerification(cfg))
			authGroup.POST("/user/verify-email", admin.VerifyEmail)

			// Admin (system-level) privilege endpoints: RequireRole("admin")
			adminRoleGroup := authGroup.Group("")
			adminRoleGroup.Use(middleware.RequireRole("admin"))
			{
				// User CRUD
				adminRoleGroup.GET("/user/list", admin.ListUsers)
				adminRoleGroup.GET("/user/detail/:id", admin.GetUser)
				adminRoleGroup.POST("/user/create", admin.CreateUser)
				adminRoleGroup.POST("/user/update", admin.UpdateUser)
				adminRoleGroup.POST("/user/delete", admin.DeleteUser)

				// Organization management
				adminRoleGroup.GET("/organizations/list", admin.ListOrganizations)
				adminRoleGroup.POST("/organizations/create", admin.CreateOrganization)
				adminRoleGroup.POST("/organizations/update", admin.UpdateOrganization)
				adminRoleGroup.POST("/organizations/delete", admin.DeleteOrganization)
			}

			// Organization-scoped endpoints: RequireOrgRole("org_admin", "org_member")
			orgMemberGroup := authGroup.Group("")
			orgMemberGroup.Use(middleware.RequireOrgRole("org_admin", "org_member"))
			{
				// Peer management
				orgMemberGroup.GET("/peer/list", admin.ListPeers)
				orgMemberGroup.GET("/peer/detail/:id", admin.GetPeer)
				orgMemberGroup.POST("/peer/delete", admin.DeletePeer)
				orgMemberGroup.POST("/peer/batchDelete", admin.BatchDeletePeer)
				orgMemberGroup.POST("/peer/update", admin.UpdatePeer)

				// Address book
				orgMemberGroup.GET("/address_book/list", admin.ListAddressBooks)
				orgMemberGroup.GET("/address_book/detail/:id", admin.GetAddressBook)
				orgMemberGroup.POST("/address_book/create", admin.CreateAddressBook)
				orgMemberGroup.POST("/address_book/update", admin.UpdateAddressBook)
				orgMemberGroup.POST("/address_book/delete", admin.DeleteAddressBook)

				// Tags
				orgMemberGroup.GET("/tag/list", admin.ListTags)
				orgMemberGroup.GET("/tag/detail/:id", admin.GetTag)
				orgMemberGroup.POST("/tag/create", admin.CreateTag)
				orgMemberGroup.POST("/tag/update", admin.UpdateTag)
				orgMemberGroup.POST("/tag/delete", admin.DeleteTag)
			}

			// Device groups: RequireOrgRole("org_admin")
			orgAdminGroup := authGroup.Group("")
			orgAdminGroup.Use(middleware.RequireOrgRole("org_admin"))
			{
				orgAdminGroup.GET("/device_group/list", admin.ListDeviceGroups)
				orgAdminGroup.GET("/device_group/detail/:id", admin.GetDeviceGroup)
				orgAdminGroup.POST("/device_group/create", admin.CreateDeviceGroup)
				orgAdminGroup.POST("/device_group/update", admin.UpdateDeviceGroup)
				orgAdminGroup.POST("/device_group/delete", admin.DeleteDeviceGroup)
			}

			// Organization user management: RequireOrgRole("org_admin")
			orgUserMgmtGroup := authGroup.Group("/organizations/:orgID/users")
			orgUserMgmtGroup.Use(middleware.RequireOrgRole("org_admin"))
			{
				orgUserMgmtGroup.GET("/list", admin.ListOrganizationUsers)
				orgUserMgmtGroup.POST("/add", admin.AddUserToOrganization)
				orgUserMgmtGroup.POST("/remove", admin.RemoveUserFromOrganization)
				orgUserMgmtGroup.POST("/update-role", admin.UpdateUserOrgRole)
			}

			// Audit logs: RequireOrgRole("org_admin", "org_auditor")
			orgAuditGroup := authGroup.Group("")
			orgAuditGroup.Use(middleware.RequireOrgRole("org_admin", "org_auditor"))
			{
				orgAuditGroup.GET("/audit_conn/list", admin.ListAuditConns)
				orgAuditGroup.GET("/audit_file/list", admin.ListAuditFiles)
			}

			// Login logs: RequireRole("admin", "auditor")
			loginLogGroup := authGroup.Group("")
			loginLogGroup.Use(middleware.RequireRole("admin", "auditor"))
			{
				loginLogGroup.GET("/login_log/list", admin.ListLoginLogs)
			}
		}
	}

	// === WebSocket ===
	hub := wsutil.NewHub()
	go hub.Run()
	r.GET("/ws", wshandler.WSUpgradeHandler(hub))
}

// versionHandler returns the API server version.
func versionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": "0.1.0",
		"name":    "RustDesk API Server",
	})
}

// healthHandler returns a simple health check response.
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// notImplemented is a placeholder handler that returns 501 Not Implemented.
func notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "This endpoint is not yet implemented",
	})
}
