package api_v1

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/krystal/krystal-network-tools/backend/ratelimiter"
	"go.uber.org/zap"
)

// Init initializes the API.
func Init(g *gin.RouterGroup, log *zap.Logger) {
	userIp(g)
	ping(g.Group("/ping", ratelimiter.NewBucket(log, 75, time.Minute, time.Minute*10)))
	dns(g.Group("/dns", ratelimiter.NewBucket(log, 20, time.Hour, time.Hour)), log)
	traceroute(g.Group("/traceroute"), log)
	bgp(g.Group("/bgp"))
	whois(g.Group("/whois"))
	rdns(g.Group("/rdns"))
}
