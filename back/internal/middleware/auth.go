package middleware

import (
	"net/http"
	"strings"

	"qabulgo-backend/internal/config"
	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	UserIDContextKey = "userID"
	RoleContextKey   = "userRole"
)

// Auth checks and validates the Bearer JWT access token
func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse(
				"Avtorizatsiya talab qilinadi. Token taqdim etilmagan.",
				nil,
			))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse(
				"Token formati noto‘g‘ri. 'Bearer <token>' ko‘rinishida bo‘lishi kerak.",
				nil,
			))
			return
		}

		tokenStr := parts[1]
		claims, err := utils.ValidateAccessToken(tokenStr, cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse(
				"Token yaroqsiz yoki muddati tugagan.",
				nil,
			))
			return
		}

		// Save in Gin Context
		c.Set(UserIDContextKey, claims.UserID)
		c.Set(RoleContextKey, models.UserRole(claims.Role))
		c.Next()
	}
}

// RequireRole enforces specific user roles for protected routes
func RequireRole(allowedRoles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(RoleContextKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.NewErrorResponse(
				"Foydalanuvchi roli aniqlanmadi.",
				nil,
			))
			return
		}

		userRole, ok := roleVal.(models.UserRole)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.NewErrorResponse(
				"Foydalanuvchi roli noto‘g‘ri shaklda.",
				nil,
			))
			return
		}

		allowed := false
		for _, r := range allowedRoles {
			if userRole == r {
				allowed = true
				break
			}
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.NewErrorResponse(
				"Ushbu amalni bajarish uchun sizda yetarli ruxsat yo‘q.",
				nil,
			))
			return
		}

		c.Next()
	}
}

// GetCurrentUserID extracts the authenticated user's UUID from Gin Context
func GetCurrentUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(UserIDContextKey)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}
