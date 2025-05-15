package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/caddyserver/certmagic"
	"github.com/gin-gonic/gin"
	"github.com/krystal/krystal-network-tools/backend/api_v1"
	"github.com/krystal/krystal-network-tools/backend/dns"
	"github.com/krystal/krystal-network-tools/backend/frontend"
	"github.com/krystal/krystal-network-tools/backend/services"
	"github.com/spf13/cobra"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Register http command to the root command
func init() {
	rootCmd.AddCommand(httpCmd)
}

// httpService is an interface that defines the methods for the HTTP server
type httpService interface {
	Run() error
	Shutdown(ctx context.Context) error
}

// httpServer is a wrapper around the http.Server
type httpServer struct {
	// The HTTP server
	*http.Server
}

// Run starts the HTTP server
// If TLSConfig is nil, it starts an HTTP server.
// If TLSConfig is not nil, it starts both HTTP and HTTPS servers and redirects HTTP -> HTTPS.
func (s *httpServer) Run() error {
	if s.TLSConfig == nil {
		slog.Info("Serving HTTP on", "addr", s.Addr)
		return s.ListenAndServe()
	}

	return certmagic.HTTPS(httpHosts, s.Handler)
}

// Shutdown gracefully shuts down the HTTP server
func (s *httpServer) Shutdown(ctx context.Context) error {
	if s.TLSConfig == nil {
		return s.Server.Shutdown(ctx)
	}

	return nil
}

// httpCmd is the command that starts the HTTP server
// It is registered to the root command and can be run with the `http` command.
// r is the gin router that serves the frontend application
// httpHosts is the list of hosts that the HTTP server listens on
// tlsConfig is the TLS configuration for the HTTPS server
var (
	httpCmd = &cobra.Command{
		Use:   "http",
		Short: "Starts the HTTP server",
		Long: `Starts the HTTP server for the Krystal Network Tools. 
			This server provides an API for interacting with the tools.`,
		Run:    runHttpService,
		PreRun: preHttpServiceRun,
	}

	r *gin.Engine

	httpHosts []string
	tlsConfig *tls.Config
)

// httpService is the main function that starts the HTTP server
func runHttpService(cmd *cobra.Command, args []string) {

	// Get the port from the environment variable or use the default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := httpServer{
		&http.Server{
			Addr:      fmt.Sprintf(":%s", port),
			Handler:   r,
			TLSConfig: tlsConfig,
		},
	}

	go func() {
		if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to start HTTPS server", "error", err)
			os.Exit(1)
		}
	}()

	// Start the pinger
	services.StartPinger(context.Background())
	slog.Info("Pinger started")

	services.GetPinger().Logf = func(s string, i ...interface{}) {
		slog.With("service", "pinger").Info(fmt.Sprintf(s, i...))
	}

	k := make(chan os.Signal, 1)
	signal.Notify(k, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-k

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slog.Info("Attempting to gracefully shutdown HTTP server...")
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Failed to gracefully shutdown HTTP server", "error", err)
		os.Exit(1)
	}
	slog.Info("HTTP server stopped")
	os.Exit(0)
}

// preHttpServiceRun Ensures that all necessary assets are loaded before starting the HTTP server1
func preHttpServiceRun(cmd *cobra.Command, args []string) {
	if os.Getenv("APP_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	r = gin.New()
	// Add default middlewares
	r.Use(gin.Recovery())
	// Add routes
	frontendRouter(r.Group("/"))
	backendRouter(r.Group("/v1"))
	// Get the HTTP host from the environment variable
	host := os.Getenv("HTTPS_HOST")
	// When the host is empty, listen for X-Forwarded-For.
	// This is useful when running behind a proxy or a Load Balancer.
	// Return if the host is empty.
	if host == "" {
		r.ForwardedByClientIP = true
		return
	}

	err := r.SetTrustedProxies(nil)
	if err != nil {
		slog.Error("Failed to set trusted proxies", "error", err)
		os.Exit(1)
	}

	// Create empty TLS config
	tlsConfig = &tls.Config{}
}

// frontendRouter is a gin router that serves the frontend application
func frontendRouter(r gin.IRouter) {
	frontend.InitFrontend(r)
}

// backendRouter is a gin router that serves the backend API
func backendRouter(r gin.IRouter) {
	// Add backend middlewares
	r.Use(cors, json, errorHandling)
	// Add routes for the backend API
	api_v1.Init(r, dns.GetCachedDNSServer(), services.GetPinger())
}

// errorHandling is a middleware that handles errors
func errorHandling(c *gin.Context) {
	c.Next()
	if len(c.Errors) > 0 {
		if c.Errors[0].Type == gin.ErrorTypePublic {
			c.JSON(c.Writer.Status(), gin.H{
				"message": c.Errors[0].Error(),
			})
		}
	}
}

// cors is a middleware that adds CORS headers to the response
func cors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
	}
}

// json is a middleware that sets the Content-Type header to application/json
// and checks if the Content-Type header is set to application/json
// If not, it returns a 400 Bad Request error
func json(c *gin.Context) {
	c.SetAccepted("application/json")
	c.Header("Content-Type", "application/json")
	c.Header("Accept", "application/json")

	if c.Request.Header.Get("Content-Type") != "application/json" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Content-Type must be application/json",
		})
		c.Abort()
		return
	}

	c.Next()
}
