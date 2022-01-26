package api_v1

import (
	"net"

	"github.com/gin-gonic/gin"
)

func rdns(g *gin.RouterGroup) {
	g.GET("/:ip", func(ctx *gin.Context) {
		ip := ctx.Param("ip")
		if hosts, _ := net.LookupAddr(ip); hosts != nil && len(hosts) > 0 {
			if ctx.ContentType() == "application/json" {
				ctx.JSON(200, map[string]string{
					"hostname": hosts[0],
				})
			} else {
				ctx.String(200, hosts[0])
			}
			return
		}
		if ctx.ContentType() == "application/json" {
			ctx.JSON(400, map[string]string{
				"message": "Failed to find IP",
			})
		} else {
			ctx.String(400, "Failed to find IP")
		}
	})
}
