package main

import (
	"os"
	"time"

	"github.com/caddyserver/certmagic"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	api "github.com/krystal/krystal-network-tools/backend/api_v1"
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

	// Handle trusted proxies.
	if err := r.SetTrustedProxies(nil); err != nil {
		logger.Fatal("Failed to set trusted proxies", zap.Error(err))
	}

	// Add the rest of the middleware/routes.
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))
	g := r.Group("/v1")
	api.Init(g, logger)

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
