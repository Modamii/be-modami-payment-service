package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modami/be-payment-service/pkg/cache"
	apperrors "github.com/modami/be-payment-service/pkg/errors"
)

// RateLimitMiddleware limits requests per user per endpoint bucket.
// maxRequests is allowed per window duration.
func RateLimitMiddleware(redisClient *cache.Client, bucket string, maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		if userID == nil {
			userID = c.ClientIP()
		}

		key := fmt.Sprintf("ratelimit:%s:%s:%d", bucket, userID, time.Now().Unix()/int64(window.Seconds()))

		count, err := redisClient.Incr(c.Request.Context(), key)
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			// Set TTL only on first increment.
			_ = redisClient.Expire(c.Request.Context(), key, window)
		}

		if count > int64(maxRequests) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    apperrors.ErrRateLimited.Code,
					"message": apperrors.ErrRateLimited.Message,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
