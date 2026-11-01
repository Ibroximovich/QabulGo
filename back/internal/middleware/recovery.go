package middleware

import (
	"log"
	"net/http"

	"qabulgo-backend/internal/dto"

	"github.com/gin-gonic/gin"
)

// Recovery handles panics and returns a generic 500 JSON response without leaking internal errors
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				reqID, _ := c.Get(RequestIDKey)
				log.Printf("[PANIC RECOVERED] RequestID=%v Error=%v", reqID, r)

				c.AbortWithStatusJSON(http.StatusInternalServerError, dto.NewErrorResponse(
					"Ichki server xatoligi yuz berdi. Iltimos, keyinroq qayta urinib ko‘ring.",
					nil,
				))
			}
		}()
		c.Next()
	}
}
