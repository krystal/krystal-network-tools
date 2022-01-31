package main

import (
	"os"
	"time"

	"github.com/caddyserver/certmagic"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	api "github.com/krystal/krystal-network-tools/backend/api_v1"
	"github.com/krystal/krystal-network-tools/backend/dns"
	"go.uber.org/zap"
)

func main() {
	// Make a zap logger.
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	// Make the gin server.
	r := gin.New()

	// Handle CORS.
	r.Use(func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type")

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
		}
	})

	// Handle internal server errors.
	r.Use(func(ctx *gin.Context) {
		ctx.Next()
		if len(ctx.Errors) != 0 {
			ferr := ctx.Errors[0]
			if ferr.Type == gin.ErrorTypePublic {
				if ctx.ContentType() == "application/json" {
					ctx.JSON(400, map[string]string{
						"message": ferr.Error(),
					})
				} else {
					ctx.String(400, ferr.Error())
				}
			} else {
				ctx.String(500, "Internal Server Error")
				logger.Error("internal server error", zap.Error(ctx.Errors[0]))
			}
		}
	})

	// Handle trusted proxies.
	if err := r.SetTrustedProxies(nil); err != nil {
		logger.Fatal("Failed to set trusted proxies", zap.Error(err))
	}

	// Add the rest of the middleware/routes.
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(gin.Recovery())
	//r.Use(ginzap.RecoveryWithZap(logger, true))
	g := r.Group("/v1")
	api.Init(g, logger, dns.GetCachedDNSServer(logger))

	// Build the listener.
	httpsHost := os.Getenv("HTTPS_HOST")
	if httpsHost == "" {
		if err = r.Run(); err != nil {
			logger.Fatal("Failed to run the server", zap.Error(err))
		}
	} else {
		if err = certmagic.HTTPS([]string{httpsHost}, r); err != nil {
			logger.Fatal("Failed to run the server", zap.Error(err))
		}
	}
}
