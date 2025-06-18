package api_v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	dnsLib "github.com/krystal/krystal-network-tools/backend/dns"
)

type dnsParams struct {
	// Trace is used to define if the DNS record should be traced all the way to the nameserver.
	Trace bool `form:"trace"`
}

func dns(g *gin.RouterGroup, dnsServer string) {
	g.GET("/:recordType/:hostname", func(ginCtx *gin.Context) {
		// Bind the params.
		var params dnsParams
		if err := ginCtx.BindQuery(&params); err != nil {
			ginCtx.JSON(http.StatusBadRequest, map[string]string{
				"message": err.Error(),
			}) // test

			return
		}

		// Get the type and hostname from the URL.
		recordType := ginCtx.Param("recordType")
		hostname := strings.TrimSuffix(ginCtx.Param("hostname"), ".")
		if hostname == "" {
			ginCtx.Error(&gin.Error{
				Type: gin.ErrorTypePublic,
				Err:  errors.New("invalid hostname"),
			})
			return
		}

		// Do the DNS lookup.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		results, err := dnsLib.LookupWithContext(ctx, dnsServer, recordType, hostname, params.Trace)
		if err != nil && results == nil {
			ginCtx.Error(&gin.Error{
				Type: gin.ErrorTypePublic,
				Err:  fmt.Errorf("failed to perform dns lookup: %v", err),
			})
			return
		}

		// If the results are empty, we still want to return a 200 OK with an empty array.
		if err != nil {
			// Return partial results with error message
			ginCtx.JSON(http.StatusPartialContent, gin.H{
				"results": results,
				"error":   err.Error(),
			})
			return
		}

		ginCtx.JSON(http.StatusOK, results)
	})
}
