package api_v1

import (
	"net"
	"time"

	"github.com/gin-gonic/gin"
	traceroutego "github.com/pixelbender/go-traceroute/traceroute"
	"go.uber.org/zap"
)

type tracerouteParams struct {
	Timeout uint `form:"timeout"`
	TTL     uint `form:"ttl"`
}

func traceroute(g *gin.RouterGroup, log *zap.Logger) {
	g.GET("/:hostnameOrIp", func(ctx *gin.Context) {
		// Get the hostname or IP.
		hostnameOrIp := ctx.Param("hostnameOrIp")

		// Parse the traceroute parameters.
		var p tracerouteParams
		if err := ctx.BindQuery(&p); err != nil {
			ctx.String(400, "unable to parse query params: %s", err.Error())
			return
		}

		// Get the addresses.
		addrs, err := net.LookupHost(hostnameOrIp)
		if err != nil {
			ctx.String(400, "invalid hostname or IP")
			return
		}
		ipAddr := net.ParseIP(addrs[0])

		// Build the tracer.
		t := &traceroutego.Tracer{
			Config: traceroutego.Config{
				Timeout:  time.Millisecond * 100,
				Delay:    time.Duration(1),
				MaxHops:  1,
				Networks: []string{"ip4:icmp", "ip4:ip", "ip6:icmp", "ip6:ip"},
			},
		}
		defer t.Close()

		// Make the session.
		s, err := t.NewSession(ipAddr)
		if err != nil {
			ctx.String(500, "Internal Server Error")
			log.Error("failed to start traceroute session", zap.Error(err))
			return
		}
		if p.Timeout == 0 || p.Timeout > 1000 {
			p.Timeout = 1000
		}
		ttl := int(p.TTL)
		if ttl == 0 {
			ttl = 64
		}
		if err = s.Ping(ttl); err != nil {
			ctx.String(500, "Internal Server Error")
			log.Error("failed to ping host", zap.Error(err))
			return
		}

		// Handle the timeout.
		// TODO: format this better, handle multi hop formatting
		select {
		case r := <-s.Receive():
			ctx.String(200, r.IP.String())
		case <-time.After(time.Millisecond * time.Duration(p.Timeout)):
			ctx.String(400, "trace timeout")
		}
	})
}
