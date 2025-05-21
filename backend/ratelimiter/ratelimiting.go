package ratelimiter

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hako/durafmt"
)

type ipBucket struct {
	backoffUntil time.Time
	reqs         uint64
}

type bucketContext interface {
	ClientIP() string
	ContentType() string
	FullPath() string
	JSON(status int, obj interface{})
	String(status int, format string, values ...interface{})
	Abort()
	Value(key interface{}) interface{}
}

type bucketTimers struct {
	sync.Mutex
	timers []*time.Timer
}

func (b *bucketTimers) add(t *time.Timer) {
	b.Lock()
	b.timers = append(b.timers, t)
	b.Unlock()
}

// Creates a new bucket. Is internal to allow for testing.
func newBucket(maxUses uint64, per, backoff time.Duration) func(bucketContext) {
	m := map[string]*ipBucket{}
	mu := sync.Mutex{}
	return func(c bucketContext) {
		mu.Lock()
		defer mu.Unlock()

		clientIp := c.ClientIP()
		b, ok := m[clientIp]
		if !ok {
			b = &ipBucket{}
			m[clientIp] = b
		}

		backoffUntilZero := b.backoffUntil.IsZero()
		if !backoffUntilZero || b.reqs == maxUses {
			if backoffUntilZero {
				b.backoffUntil = time.Now().Add(backoff)
				time.AfterFunc(backoff, func() {
					mu.Lock()
					delete(m, clientIp)
					mu.Unlock()
				})
			}

			slog.With("client_ip", clientIp, "handler_path", c.FullPath(), "path", c.Value(0).(*http.Request).URL.Path).
				Warn("ratelimited user trying request")

			durationFmt := "forever"
			var duration time.Duration
			if maxUses != 0 && per != 0 {
				duration = b.backoffUntil.Sub(time.Now())
				durationFmt = durafmt.Parse(duration).LimitFirstN(2).String()
			}

			if c.ContentType() == "application/json" {
				c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"wait_ms": duration.Milliseconds(),
					"message": "You have been ratelimited! Try again later.",
				})
			} else {
				c.String(http.StatusTooManyRequests, "You have been ratelimited! Try again in %s.", durationFmt)
			}

			c.Abort()
			return
		}

		x := b.reqs
		b.reqs++

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

// NewBucket is used to create a new ratelimit bucket for users of the site.
func NewBucket(maxUses uint64, per, backoff time.Duration) gin.HandlerFunc {
	b := newBucket(maxUses, per, backoff)
	return func(context *gin.Context) {
		b(context)
	}
}
