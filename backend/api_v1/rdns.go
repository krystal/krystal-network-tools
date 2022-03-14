package api_v1

import (
	"fmt"
	"net"

	"github.com/gin-gonic/gin"
	dnsLib "github.com/krystal/krystal-network-tools/backend/dns"
	"go.uber.org/zap"
)

type rdnsParams struct {
	// Trace is used to define if the DNS record should be traced all the way to the nameserver.
	Trace bool `form:"trace"`
}

func rdns(g group, log *zap.Logger, dnsServer string) {
	g.GET("/:ip", func(ctx *gin.Context) {
		ip := ctx.Param("ip")

		isJson := ctx.ContentType() == "application/json"
		var params rdnsParams
		if err := ctx.BindQuery(&params); err != nil {
			if isJson {
				ctx.JSON(400, map[string]string{
					"message": err.Error(),
				})
			} else {
				ctx.String(400, "unable to parse query params: %s", err.Error())
			}
			return
		}

		if !params.Trace {
			hosts, err := net.LookupAddr(ip)
			if err != nil || len(hosts) == 0 {
				if ctx.ContentType() == "application/json" {
					ctx.JSON(400, map[string]string{
						"message": "Failed to find IP",
					})
				} else {
					ctx.String(400, "Failed to find IP")
				}
				return
			}
			if ctx.ContentType() == "application/json" {
				ctx.JSON(200, map[string]string{
					"hostname": hosts[0],
				})
			} else {
				ctx.String(200, hosts[0])
			}

			return
		}

		ipAddr := net.ParseIP(ip)
		result, err := dnsLib.LookupRDNS(
			log, ipAddr, dnsServer,
		)
		if err != nil {
			ctx.Error(&gin.Error{
				Type: gin.ErrorTypePublic,
				Err:  fmt.Errorf("failed to perform dns lookup: %v", err),
			})
			return
		}
		if ctx.ContentType() == "application/json" {
			ctx.JSON(200, map[string]interface{}{
				"trace": result,
			})
		} else {
			ctx.String(200, result.String())
		}
	})
}
