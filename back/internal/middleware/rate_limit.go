package middleware

import (
	"net/http"
	"sync"
	"time"

	"qabulgo-backend/internal/dto"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	mu      sync.RWMutex
	clients map[string]*clientLimiter
	r       rate.Limit
	b       int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		clients: make(map[string]*clientLimiter),
		r:       r,
		b:       b,
	}

	// Clean up stale clients every 5 minutes
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			i.mu.Lock()
			for ip, client := range i.clients {
				if time.Since(client.lastSeen) > 10*time.Minute {
					delete(i.clients, ip)
				}
			}
			i.mu.Unlock()
		}
	}()

	return i
}

func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	client, exists := i.clients[ip]
	if !exists {
		limiter := rate.NewLimiter(i.r, i.b)
		i.clients[ip] = &clientLimiter{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	client.lastSeen = time.Now()
	return client.limiter
}

// RateLimit limits requests per IP address
func RateLimit(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.getLimiter(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, dto.NewErrorResponse(
				"So‘rovlar soni me’yordan oshib ketdi. Iltimos, birozdan so‘ng qayta urinib ko‘ring.",
				nil,
			))
			return
		}
		c.Next()
	}
}
