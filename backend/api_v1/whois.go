package api_v1

import (
	"github.com/gin-gonic/gin"
	gowhois "github.com/likexian/whois"
)

func whois(g *gin.RouterGroup) {
	g.GET("/:hostOrIp", func(context *gin.Context) {
		// Get the hostname or IP address.
		hostOrIp := context.Param("hostOrIp")

		// Do the WHOIS lookup.
		result, err := gowhois.Whois(hostOrIp)
		if err != nil {
			if context.ContentType() == "application/json" {
				context.JSON(400, map[string]string{
					"message": err.Error(),
				})
			} else {
				context.String(400, "WHOIS lookup failed: %s", err.Error())
			}
			return
		}

		// Return a 200.
		if context.ContentType() == "application/json" {
			context.JSON(200, map[string]string{
				"result": result,
			})
		} else {
			context.String(200, result)
		}
	})
}
