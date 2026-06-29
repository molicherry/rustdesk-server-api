package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ServeEmbeddedFrontend registers routes to serve the embedded React SPA.
// frontendFS must be an embed.FS rooted at the project directory containing
// web/dist/ with the built frontend assets.
//   - Static files (favicon, icons, assets) are served from the embedded filesystem.
//   - A NoRoute handler catches all unmatched paths and returns index.html
//     for client-side routing (SPA fallback). API paths (/api/, /ws) are
//     excluded from the fallback and return a standard 404 JSON response.
func ServeEmbeddedFrontend(r *gin.Engine, frontendFS embed.FS) {
	subFS, err := fs.Sub(frontendFS, "web/dist")
	if err != nil {
		logrus.WithError(err).Error("failed to create sub filesystem from embedded frontend")
		return
	}
	fsys := http.FS(subFS)

	fileServer := http.FileServer(fsys)

	// Serve static files at their exact paths.
	// These must be registered before NoRoute so Gin matches them first.
	staticFiles := []string{
		"/favicon.svg",
		"/icons.svg",
	}
	for _, path := range staticFiles {
		p := path
		r.GET(p, func(c *gin.Context) {
			c.Request.URL.Path = p
			fileServer.ServeHTTP(c.Writer, c.Request)
		})
	}

	// Serve the assets directory (bundled JS/CSS).
	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.Request.URL.Path = "/assets/" + c.Param("filepath")
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	// SPA fallback: any route that doesn't match a registered API route
	// returns index.html so React Router can handle client-side navigation.
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Never fallback for API or WebSocket routes — those are real 404s.
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/ws") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		// If the requested path exists as a regular file, serve it directly.
		// Skip directories (including root "/") to avoid http.FileServer
		// 301 redirects and fall through to the SPA fallback below.
		trimmed := strings.TrimPrefix(path, "/")
		f, err := fsys.Open(trimmed)
		if err == nil {
			info, statErr := f.Stat()
			f.Close()
			if statErr == nil && !info.IsDir() {
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}

		// SPA fallback: serve index.html for client-side routing.
		// Read the file directly and set the response body rather than
		// delegating to http.FileServer, which can trigger unwanted 301
		// redirects when the request path does not match a real file.
		idx, err := fs.ReadFile(subFS, "index.html")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", idx)
	})

	logrus.Info("embedded frontend enabled (web_client=true)")
}
