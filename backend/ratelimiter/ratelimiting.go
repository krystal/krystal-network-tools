package ratelimiter

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hako/durafmt"
	"go.uber.org/zap"
)

type bucket struct {
	backoffUntil time.Time
	reqs         uint64
}

// NewBucket is used to create a new ratelimit bucket for users of the site.
func NewBucket(log *zap.Logger, maxUses uint64, per, backoff time.Duration) gin.HandlerFunc {
	m := map[string]*bucket{}
	mu := sync.Mutex{}
	return func(c *gin.Context) {
		// Lock the global map whilst we handle this. We will almost always write.
		mu.Lock()

		// Defer unlocking the global map to make sure we don't deadlock it.
		defer mu.Unlock()

		// Get the list and if it doesn't exist make it.
		clientIp := c.ClientIP()
		b, ok := m[clientIp]
		if !ok {
			b = &bucket{}
			m[clientIp] = b
		}

		// Check if we have exceeded the maximum number of uses.
		backoffUntilZero := b.backoffUntil.IsZero()
		if !backoffUntilZero || b.reqs == maxUses {
			// If we do not have a backoff time set, set it.
			if backoffUntilZero {
				b.backoffUntil = time.Now().Add(backoff)
				time.AfterFunc(backoff, func() {
					mu.Lock()
					delete(m, clientIp)
					mu.Unlock()
				})
			}

			// Log a warning since this is potential abuse.
			log.Warn("ratelimited user trying request", zap.String("client_ip", clientIp),
				zap.String("handler_path", c.FullPath()), zap.String("path", c.Request.URL.Path))

			// Get the first timestamp.
			durationFmt := "forever"
			var duration time.Duration
			if maxUses != 0 && per != 0 {
				duration = b.backoffUntil.Sub(time.Now())
				durationFmt = durafmt.Parse(duration).LimitFirstN(2).String()
			}

			// Return an error to the user and return.
			if c.ContentType() == "application/json" {
				c.JSON(429, map[string]interface{}{
					"wait_ms": duration.Milliseconds(),
					"message": fmt.Sprintf("You have been ratelimited! Try again in %s.", durationFmt),
				})
			} else {
				c.String(429, "You have been ratelimited! Try again in %s.", durationFmt)
			}
			c.Abort()
			return
		}

		// Add 1 to the bucket.
		x := b.reqs
		b.reqs++

		// Create a function to zero the request count after the per duration on the first request of the bucket.
		if x == 0 {
			time.AfterFunc(per, func() {
				mu.Lock()
				b.reqs = 0
				if b.backoffUntil.IsZero() {
					delete(m, clientIp)
				}
				mu.Unlock()
			})
		}
	}
}