package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// Localization is a middleware that parses the Accept-Language header
// and sets the preferred locale in the gin context.
// It supports en and zh-CN. Falls back to en.
func Localization() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := parseAcceptLanguage(c.GetHeader("Accept-Language"))
		c.Set("locale", locale)
		c.Next()
	}
}

// parseAcceptLanguage extracts the first supported locale from an Accept-Language header.
// Example input: "zh-CN,zh;q=0.9,en;q=0.8"
// Returns "zh-CN" if zh-CN or zh is present, otherwise "en".
func parseAcceptLanguage(header string) string {
	if header == "" {
		return "en"
	}
	for _, part := range strings.Split(header, ",") {
		lang := strings.TrimSpace(part)
		// Strip quality value
		if idx := strings.Index(lang, ";"); idx != -1 {
			lang = lang[:idx]
		}
		lang = strings.TrimSpace(lang)
		switch {
		case strings.HasPrefix(strings.ToLower(lang), "zh-cn"),
			strings.HasPrefix(strings.ToLower(lang), "zh-hans"),
			strings.EqualFold(lang, "zh"):
			return "zh-CN"
		}
	}
	return "en"
}
