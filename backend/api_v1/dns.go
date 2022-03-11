package api_v1

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	dnsLib "github.com/krystal/krystal-network-tools/backend/dns"
	"go.uber.org/zap"
)

type dnsParams struct {
	// Trace is used to define if the DNS record should be traced all the way to the nameserver.
	Trace bool `form:"trace"`
}

func dns(g *gin.RouterGroup, log *zap.Logger, dnsServer string) {
	g.GET("/:recordType/:hostname", func(context *gin.Context) {
		// Defines if this is JSON.
		isJson := context.ContentType() == "application/json"

		// Bind the params.
		var params dnsParams
		if err := context.BindQuery(&params); err != nil {
			if isJson {
				context.JSON(400, map[string]string{
					"message": err.Error(),
				})
			} else {
				context.String(400, "unable to parse query params: %s", err.Error())
			}
			return
		}

		// Get the type and hostname from the URL.
		recordType := context.Param("recordType")
		hostname := strings.TrimSuffix(context.Param("hostname"), ".")
		if hostname == "" {
			context.Error(&gin.Error{
				Type: gin.ErrorTypePublic,
				Err:  errors.New("invalid hostname"),
			})
			return
		}

		// Do the DNS lookup.
		results, err := dnsLib.Lookup(
			log, dnsServer, recordType, hostname, params.Trace,
		)
		if err != nil {
			context.Error(&gin.Error{
				Type: gin.ErrorTypePublic,
				Err:  fmt.Errorf("failed to perform dns lookup: %v", err),
			})
			return
		}

		// Handle JSON responses.
		if isJson {
			context.JSON(200, results)
			return
		}

		context.String(200, results.String())
	})
}
