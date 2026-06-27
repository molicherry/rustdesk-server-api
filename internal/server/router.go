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
func RegisterRoutes(r *gin.Engine, _ *config.Config) {
	// Health check
	r.GET("/api/version", versionHandler)
	r.GET("/api/health", healthHandler)

	// === Client API group (/api/*) ===
	apiGroup := r.Group("/api")
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
		adminGroup.POST("/logout", admin.AdminLogout)
		adminGroup.GET("/captcha", admin.Captcha)
		adminGroup.POST("/user/register", admin.UserRegister)
		adminGroup.GET("/config/server", admin.ConfigServer)

		// Authenticated endpoints (BackendUserAuth required)
		authGroup := adminGroup.Group("")
		authGroup.Use(middleware.BackendUserAuth())
		{
			authGroup.GET("/user/current", admin.GetCurrentUser)
			authGroup.POST("/user/changeCurPwd", admin.ChangeCurrentPassword)

			// Admin privilege endpoints (BackendUserAuth + AdminPrivilege)
			adminPrivGroup := authGroup.Group("")
			adminPrivGroup.Use(middleware.AdminPrivilege())
			{
				// User CRUD
				adminPrivGroup.GET("/user/list", admin.ListUsers)
				adminPrivGroup.GET("/user/detail/:id", admin.GetUser)
				adminPrivGroup.POST("/user/create", admin.CreateUser)
				adminPrivGroup.POST("/user/update", admin.UpdateUser)
				adminPrivGroup.POST("/user/delete", admin.DeleteUser)

				// Peer management (Phase 2)
				adminPrivGroup.GET("/peer/list", admin.ListPeers)
				adminPrivGroup.GET("/peer/detail/:id", admin.GetPeer)
				adminPrivGroup.POST("/peer/delete", admin.DeletePeer)
				adminPrivGroup.POST("/peer/batchDelete", admin.BatchDeletePeer)
				adminPrivGroup.POST("/peer/update", admin.UpdatePeer)

				// Address book (Phase 3)
				adminPrivGroup.GET("/address_book/list", admin.ListAddressBooks)
				adminPrivGroup.GET("/address_book/detail/:id", admin.GetAddressBook)
				adminPrivGroup.POST("/address_book/create", admin.CreateAddressBook)
				adminPrivGroup.POST("/address_book/update", admin.UpdateAddressBook)
				adminPrivGroup.POST("/address_book/delete", admin.DeleteAddressBook)

				// Audit logs (Phase 3)
				adminPrivGroup.GET("/audit_conn/list", admin.ListAuditConns)
				adminPrivGroup.GET("/audit_file/list", admin.ListAuditFiles)

				// Login logs (Phase 3)
				adminPrivGroup.GET("/login_log/list", admin.ListLoginLogs)

				// Tags (Phase 3)
				adminPrivGroup.GET("/tag/list", admin.ListTags)
				adminPrivGroup.GET("/tag/detail/:id", admin.GetTag)
				adminPrivGroup.POST("/tag/create", admin.CreateTag)
				adminPrivGroup.POST("/tag/update", admin.UpdateTag)
				adminPrivGroup.POST("/tag/delete", admin.DeleteTag)

				// Device groups (Phase 3)
				adminPrivGroup.GET("/device_group/list", admin.ListDeviceGroups)
				adminPrivGroup.GET("/device_group/detail/:id", admin.GetDeviceGroup)
				adminPrivGroup.POST("/device_group/create", admin.CreateDeviceGroup)
				adminPrivGroup.POST("/device_group/update", admin.UpdateDeviceGroup)
				adminPrivGroup.POST("/device_group/delete", admin.DeleteDeviceGroup)
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
