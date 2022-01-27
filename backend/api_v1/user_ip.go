package api_v1

import "github.com/gin-gonic/gin"

func userIp(g *gin.RouterGroup) {
	g.GET("/ip", func(ctx *gin.Context) {
		if ctx.ContentType() == "application/json" {
			ctx.JSON(200, map[string]string{"ip": ctx.ClientIP()})
		} else {
			ctx.String(200, ctx.ClientIP())
		}
	})
}
