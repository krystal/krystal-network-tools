package api_v1

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	dnsLib "github.com/krystal/krystal-network-tools/backend/dns"
)

type dnsParams struct {
	// Trace is used to define if the DNS record should be traced all the way to the nameserver.
	Trace bool `form:"trace"`
}

func dns(g *gin.RouterGroup, dnsServer string) {
	g.GET("/:recordType/:hostname", func(context *gin.Context) {
		// Bind the params.
		var params dnsParams
		if err := context.BindQuery(&params); err != nil {
			context.JSON(http.StatusBadRequest, map[string]string{
				"message": err.Error(),
			})

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
		results, err := dnsLib.Lookup(dnsServer, recordType, hostname, params.Trace)
		if err != nil {
			context.Error(&gin.Error{
				Type: gin.ErrorTypePublic,
				Err:  fmt.Errorf("failed to perform dns lookup: %v", err),
			})
			return
		}

		context.JSON(http.StatusOK, results)

	})
}
