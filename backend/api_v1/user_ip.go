package api_v1

import "github.com/gin-gonic/gin"

func userIp(g *gin.RouterGroup) {
	g.GET("/ip", func(ctx *gin.Context) {
		ctx.String(200, ctx.ClientIP())
	})
}
