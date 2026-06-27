package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger returns a middleware that logs HTTP requests using logrus.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		entry := logrus.WithFields(logrus.Fields{
			"status":     statusCode,
			"method":     method,
			"path":       path,
			"query":      rawQuery,
			"ip":         clientIP,
			"latency":    latency.String(),
			"user-agent": c.Request.UserAgent(),
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.String())
			return
		}

		if statusCode >= 500 {
			entry.Error("server error")
		} else if statusCode >= 400 {
			entry.Warn("client error")
		} else {
			entry.Info("request")
		}
	}
}
