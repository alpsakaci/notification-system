package middleware

import (
	"net/http"

	"notification-system/internal/infrastructure/cache"

	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware limits the number of requests a client can make per second.
func RateLimitMiddleware(redisClient *cache.RedisClient, maxRequestsPerSecond int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// We can reuse the AllowRateLimit logic from RedisClient, but we should make sure the key is specific to IP
		// For simplicity, we just pass the IP as the "channel" string, since the cache implementation just concatenates it
		// e.g. "rate_limit:" + clientIP + ":" + time.Now()...

		allowed, err := redisClient.AllowRateLimit(c.Request.Context(), "api_ip:"+clientIP, maxRequestsPerSecond)
		if err != nil {
			// If redis fails, we probably shouldn't block traffic, just log it.
			// But for strict rate limiting, we could return 500. Let's allow it fallback.
			c.Next()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
