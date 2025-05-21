package api_v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func userIp(g group) {
	g.GET("/ip", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, map[string]string{"ip": ctx.ClientIP()})
	})
}
