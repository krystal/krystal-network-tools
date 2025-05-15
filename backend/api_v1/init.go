package api_v1

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/krystal/krystal-network-tools/backend/ratelimiter"
	pingttl "github.com/strideynet/go-ping-ttl"
)

// Init initializes the API.
func Init(g gin.IRouter, cachedDnsServer string, pinger *pingttl.Pinger) {
	// Create the base bucket for a few types of requests related to pinging. This works out to
	// 10 requests/second, so not awfully consequential to a server but will likely be fine for us.
	pingingBucket := ratelimiter.NewBucket(100, time.Second*10, time.Minute*10)

	// Load the routes.
	userIp(g)
	ping(g.Group("/ping", pingingBucket))
	dns(g.Group("/dns", ratelimiter.NewBucket(100, time.Hour, time.Minute*5)), cachedDnsServer)
	traceroute(g.Group("/traceroute", pingingBucket), pinger)
	bgp(g.Group("/bgp", ratelimiter.NewBucket(20, time.Hour, time.Minute*10)), makeBirdSocket)
	whois(g.Group("/whois", ratelimiter.NewBucket(20, time.Hour, time.Minute*10)), defaultWhoisLookuper{})
	rdns(g.Group("/rdns", ratelimiter.NewBucket(40, time.Hour, time.Minute*10)), cachedDnsServer)
}
