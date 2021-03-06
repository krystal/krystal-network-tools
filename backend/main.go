package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/caddyserver/certmagic"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	api "github.com/krystal/krystal-network-tools/backend/api_v1"
	"github.com/krystal/krystal-network-tools/backend/dns"
	pingttl "github.com/strideynet/go-ping-ttl"
	"go.uber.org/zap"
)

//go:embed frontend_blobs
var frontendBlobs embed.FS

type serveFsStaticImpl struct {
	http.FileSystem
}

func (i serveFsStaticImpl) Exists(prefix string, path string) bool {
	_, err := i.FileSystem.Open(path)
	return err == nil
}

var _ static.ServeFileSystem = serveFsStaticImpl{}

func errorHandler(logger *zap.Logger) func(*gin.Context) {
	return func(ctx *gin.Context) {
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
	}
}

func main() {
	// Sub the frontend blob.
	f, err := fs.Sub(frontendBlobs, "frontend_blobs")
	if err != nil {
		panic(err)
	}

	// Make a zap logger.
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	// Make the gin server.
	r := gin.New()

	// Handle initializing the frontend HTML routes.
	initFrontend(r, f, logger)

	// Add the static files.
	r.Use(static.Serve("/", serveFsStaticImpl{http.FS(f)}))

	// Handle CORS.
	r.Use(func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type")

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
		}
	})

	// Handle internal server errors.
	r.Use(errorHandler(logger))

	pinger := pingttl.New()

	go func() {
		logger.Info("starting pinger")
		err = pinger.Run(context.Background())
		if err != nil {
			logger.Fatal("failed to start pinger", zap.Error(err))
		}
	}()
	logger.Info("started pinger")

	pinger.Logf = func(s string, i ...interface{}) {
		logger.Named("pinger").Info(fmt.Sprintf(s, i...))
	}

	// Add the rest of the middleware/routes.
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))
	g := r.Group("/v1")
	api.Init(g, logger, dns.GetCachedDNSServer(logger), pinger)

	// Build the listener.
	httpsHost := os.Getenv("HTTPS_HOST")
	if httpsHost == "" {
		// Listen for X-Forwarded-For.
		r.ForwardedByClientIP = true

		// Run on the specified port.
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		ln, err := net.Listen("tcp", "127.0.0.1:"+port)
		if err != nil {
			logger.Fatal(
				"failed to listen",
				zap.Error(err),
				zap.String("port", port),
			)
		}
		logger.Info("server started", zap.String("port", port))
		if err = r.RunListener(ln); err != nil {
			logger.Fatal("Failed to run the server", zap.Error(err))
		}
	} else {
		// Handle blanking trusted proxies.
		if err := r.SetTrustedProxies(nil); err != nil {
			logger.Fatal("Failed to set trusted proxies", zap.Error(err))
		}

		// Launch with certmagic.
		if err = certmagic.HTTPS([]string{httpsHost}, r); err != nil {
			logger.Fatal("Failed to run the server", zap.Error(err))
		}
	}
}
