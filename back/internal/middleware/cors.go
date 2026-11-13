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
		addOrigin("http://localhost:3001")
		addOrigin("http://localhost:3002")
		addOrigin("http://localhost:5173")
		addOrigin("http://localhost:5174")
		addOrigin("http://localhost:5175")
		addOrigin("http://localhost:8080")
		addOrigin("http://127.0.0.1:3000")
		addOrigin("http://127.0.0.1:3001")
		addOrigin("http://127.0.0.1:3002")
		addOrigin("http://127.0.0.1:5173")
		addOrigin("http://127.0.0.1:5174")
		addOrigin("http://127.0.0.1:5175")
	}

	// Safe fallback if no valid origin was provided
	if len(allowedOrigins) == 0 {
		addOrigin("http://localhost:5173")
		addOrigin("http://localhost:3000")
	}

	allowOriginFunc := func(origin string) bool {
		if seen[origin] {
			return true
		}
		if cfg == nil || cfg.AppEnv != "production" {
			if strings.HasPrefix(origin, "http://localhost:") ||
				strings.HasPrefix(origin, "http://127.0.0.1:") ||
				origin == "http://localhost" ||
				origin == "http://127.0.0.1" {
				return true
			}
		}
		return false
	}

	corsConfig := cors.Config{
		AllowOriginFunc:  allowOriginFunc,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept", "Accept-Encoding", "Accept-Language", "Authorization", "X-Request-ID", "X-Requested-With", "Idempotency-Key", "idempotency-key", "Cache-Control", "Pragma"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID", "Idempotency-Key"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	return cors.New(corsConfig)
}
