package api_v1

import (
	"github.com/gin-gonic/gin"
	"net"
)

func rdns(g *gin.RouterGroup) {
	g.GET("/:ip", func(ctx *gin.Context) {
		ip := ctx.Param("ip")
		if hosts, _ := net.LookupAddr(ip); hosts != nil && len(hosts) > 0 {
			ctx.String(200, hosts[0])
			return
		}
		ctx.String(400, "Failed to find IP")
	})
}
