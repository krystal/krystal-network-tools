package main

import (
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	api "github.com/krystal/krystal-network-tools/backend/api_v1"
	"github.com/krystal/krystal-network-tools/backend/utils"
	"go.uber.org/zap"
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
	r := gin.Default()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	//r.Use(ginzap.RecoveryWithZap(logger, true))
	g := r.Group("/v1")
	api.Init(g, logger)
	if err = r.Run(); err != nil {
		logger.Fatal("Failed to run the server", zap.Error(err))
	}
}
