package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger logs safe details about each incoming HTTP request
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		reqID, _ := c.Get(RequestIDKey)

		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		log.Printf("[HTTP] %s | %3d | %12v | %s | %-7s %s | req_id=%v",
			time.Now().Format("2006/01/02 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
			reqID,
		)
	}
}
