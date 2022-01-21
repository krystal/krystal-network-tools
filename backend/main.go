package main

import (
	"github.com/caddyserver/certmagic"
	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	api "github.com/krystal/krystal-network-tools/backend/api_v1"
	"github.com/krystal/krystal-network-tools/backend/utils"
	"go.uber.org/zap"
	"os"
	"time"
)

func main() {
	// Make a zap logger.
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	// Initialize the server.
	utils.InitializeDNSServer(logger)

	// Make the gin server.
	r := gin.New()
	if err := r.SetTrustedProxies(nil); err != nil {
		logger.Fatal("Failed to set trusted proxies", zap.Error(err))
	}
	r.Use(cors.Default())
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
