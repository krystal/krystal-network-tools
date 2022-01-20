package api_v1

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Init initializes the API.
func Init(g *gin.RouterGroup, log *zap.Logger) {
	userIp(g)
	ping(g.Group("/ping"))
	dns(g.Group("/dns"), log)
	traceroute(g.Group("/traceroute"), log)
	bgp(g.Group("/bgp"))
	whois(g.Group("/whois"))
}
