package utils

import (
	"container/list"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hako/durafmt"
	"go.uber.org/zap"
)

// NewRatelimitBucket is used to create a new ratelimit bucket for users of the site.
// The ratelimit is rolling. This effectively means that each requests hit will expire after a certain amount of time.
func NewRatelimitBucket(log *zap.Logger, maxUses uint64, per time.Duration) gin.HandlerFunc {
	m := map[string]*list.List{}
	mu := sync.Mutex{}
	return func(c *gin.Context) {
		// Lock the global map whilst we handle this. We will almost always write.
		mu.Lock()

		// Defer unlocking the global map to make sure we don't deadlock it.
		defer mu.Unlock()

		// Get the list and if it doesn't exist make it.
		clientIp := c.ClientIP()
		l, ok := m[clientIp]
		if !ok {
			l = list.New()
			m[clientIp] = l
		}

		// Check if we have exceeded the maximum number of uses.
		if l.Len() == int(maxUses) {
			// Log a warning since this is potential abuse.
			log.Warn("ratelimited user trying request", zap.String("client_ip", clientIp),
				zap.String("handler_path", c.FullPath()), zap.String("path", c.Request.URL.Path))

			// Get the first timestamp.
			durationFmt := "forever"
			if maxUses != 0 && per != 0 {
				firstTime := l.Front().Value.(time.Time)

				// The time to unlock is calculated by the time now sub the earliest request time subtracted from
				// the ratelimit time.
				duration := per - time.Now().Sub(firstTime)

				durationFmt = durafmt.Parse(duration).LimitFirstN(2).String()
			}

			// Return an error to the user and return.
			c.String(429, "You have been ratelimited! Try again in %s.", durationFmt)
			c.Abort()
			return
		}

		// Add the current time to the list.
		e := l.PushBack(time.Now())

		// Create a function to remove the current time from the list.
		if per != 0 {
			time.AfterFunc(per, func() {
				mu.Lock()
				l.Remove(e)
				if l.Len() == 0 {
					// In this case, we should remove the list from the map.
					delete(m, clientIp)
				}
				mu.Unlock()
			})
		}
	}
}
