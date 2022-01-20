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
			context.String(400, "WHOIS lookup failed: %s", err.Error())
			return
		}

		// Return a 200.
		context.String(200, result)
	})
}
