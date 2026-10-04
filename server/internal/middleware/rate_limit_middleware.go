package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
)

// RateLimitMiddleware 是一个基于令牌桶的限流中间件
func RateLimitMiddleware(fillInterval time.Duration, capacity int64) gin.HandlerFunc {
	bucket := ratelimit.NewBucketWithRate(float64(fillInterval), capacity)

	return func(c *gin.Context) {
		if bucket.Take(1) > 0 {
			c.Next()
		} else {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
		}
	}
}
