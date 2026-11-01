package middleware

import (
	"strings"
	"time"

	"qabulgo-backend/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS configures cross-origin resource sharing policies
func CORS(cfg *config.Config) gin.HandlerFunc {
	var allowedOrigins []string
	seen := make(map[string]bool)

	addOrigin := func(origin string) {
		trimmed := strings.TrimSpace(origin)
		trimmed = strings.TrimRight(trimmed, "/")
		if trimmed == "" || trimmed == "*" {
			return
		}
		if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
			return
		}
		if !seen[trimmed] {
			seen[trimmed] = true
			allowedOrigins = append(allowedOrigins, trimmed)
		}
	}

	if cfg != nil && cfg.FrontendURL != "" {
		for _, o := range strings.Split(cfg.FrontendURL, ",") {
			addOrigin(o)
		}
	}

	// In non-production environments, allow local development ports by default
	if cfg == nil || cfg.AppEnv != "production" {
		addOrigin("http://localhost:3000")
		addOrigin("http://localhost:5173")
	}

	// Safe fallback if no valid origin was provided
	if len(allowedOrigins) == 0 {
		addOrigin("http://localhost:5173")
	}

	corsConfig := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	return cors.New(corsConfig)
}
