package api_v1

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/krystal/krystal-network-tools/backend/ratelimiter"
	pingttl "github.com/strideynet/go-ping-ttl"
	"go.uber.org/zap"
)

// Init initializes the API.
func Init(g *gin.RouterGroup, log *zap.Logger, cachedDnsServer string, pinger *pingttl.Pinger) {
	// Create the base bucket for a few types of requests.
	pingingBucket := ratelimiter.NewBucket(log, 75, time.Minute, time.Minute*10)

	// Load the routes.
	userIp(g)
	ping(g.Group("/ping", pingingBucket), log, pinger)
	dns(
		g.Group("/dns", ratelimiter.NewBucket(log, 20, time.Hour, time.Minute*10)), log,
		cachedDnsServer,
	)
	traceroute(g.Group("/traceroute", pingingBucket), log, pinger)
	bgp(g.Group("/bgp", ratelimiter.NewBucket(log, 20, time.Hour, time.Minute*10)), makeBirdSocket)
	whois(g.Group("/whois", ratelimiter.NewBucket(log, 20, time.Hour, time.Minute*10)), defaultWhoisLookuper{})
	rdns(g.Group("/rdns", ratelimiter.NewBucket(log, 40, time.Hour, time.Minute*10)))
}
