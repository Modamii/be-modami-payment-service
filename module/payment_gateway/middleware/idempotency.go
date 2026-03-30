package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modami/be-payment-service/pkg/cache"
	apperrors "github.com/modami/be-payment-service/pkg/errors"
)

const idempotencyTTL = 24 * time.Hour

// IdempotencyMiddleware checks the Idempotency-Key header and prevents duplicate processing.
// On a duplicate request it returns 409 Conflict immediately.
func IdempotencyMiddleware(redisClient *cache.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		cacheKey := "idempotency:" + key

		// Try to acquire the idempotency lock.
		set, err := redisClient.SetNX(c.Request.Context(), cacheKey, "processing", idempotencyTTL)
		if err != nil {
			// Redis error — allow the request to proceed rather than blocking.
			c.Next()
			return
		}

		if !set {
			// Key exists — duplicate request.
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error": gin.H{
					"code":    apperrors.ErrDuplicateRequest.Code,
					"message": apperrors.ErrDuplicateRequest.Message,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
