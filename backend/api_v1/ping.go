package api_v1

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	pingttl "github.com/strideynet/go-ping-ttl"
	"go.uber.org/zap"
)

type pingParams struct {
	// IPV6 should be set to true if we should try to ping via IPv6.
	IPV6 bool `form:"ipv6"`
	// Timeout in milliseconds.
	Timeout uint `form:"timeout"`
	// Count is the number of times to run a ping.
	Count uint `form:"count"`
	// Interval is the time between consecutive pings in milliseconds.
	Interval uint `form:"interval"`
}

// PingErrorMessage is used to define the error message.
type PingErrorMessage struct {
	// IsTimeout is used to define if the error is a timeout.
	IsTimeout bool `json:"is_timeout"`

	// Message is used to define the error message.
	Message string `json:"message"`
}

// PingResponse is used to define an item in the ping response.
type PingResponse struct {
	// Error is not nil if there was an error.
	Error *PingErrorMessage `json:"error,omitempty"`

	// IPAddress is used to define the IP address of the host.
	IPAddress string `json:"ip_address"`

	// Hostname is used to define the hostname. If it is nil, it means the rDNS lookup is not available.
	Hostname *string `json:"hostname"`

	// Latency is the time the round-trip in milliseconds.
	Latency *float64 `json:"latency,omitempty"`
}

type pinger interface {
	Ping(context.Context, *net.IPAddr, int) (*pingttl.PingResult, error)
}

func ping(g *gin.RouterGroup, log *zap.Logger, p pinger) {
	g.GET("/:hostnameOrIp", func(ctx *gin.Context) {
		// Get the hostname or IP.
		hostnameOrIp := ctx.Param("hostnameOrIp")

		// Defines if this is JSON.
		isJson := ctx.ContentType() == "application/json"

		// Bind the ping params.
		var params pingParams
		if err := ctx.BindQuery(&params); err != nil {
			if isJson {
				ctx.JSON(400, map[string]string{
					"message": err.Error(),
				})
			} else {
				ctx.String(400, "unable to parse query params: %s", err.Error())
			}
			return
		}

		// Enforce the maximum count.
		if params.Count > 0 {
			if params.Count > 10 {
				params.Count = 10
			}
		} else {
			params.Count = 1
		}

		lookupProtocol := "ip4"
		if params.IPV6 {
			lookupProtocol = "ip6"
		}
		addr, err := net.ResolveIPAddr(lookupProtocol, hostnameOrIp)
		if err != nil {
			ctx.Error(&gin.Error{
				Err:  errors.New("failed to resolve the ip address"),
				Type: gin.ErrorTypePublic,
			})
			log.Error("ip resolve error", zap.Error(err))
			return
		}

		// Attempt a rdns lookup.
		hostnameOrIp = addr.String()
		var hostname *string
		if hosts, _ := net.LookupAddr(hostnameOrIp); hosts != nil && len(hosts) > 0 {
			hostnameOrIp += " [" + hosts[0] + "]"
			hostname = &hosts[0]
		}

		// Defines all responses.
		strResponses := []string{}
		jsonResponses := []*PingResponse{}

		// Make sure the interval is less than or equal to 1 second.
		if params.Interval > 1000 {
			params.Interval = 1000
		}

		if params.Timeout == 0 {
			params.Timeout = 5000
		}

		for i := uint(0); i < params.Count; i++ {
			// If i isn't 0, sleep for the specified interval.
			if i != 0 {
				// Sleep for the specified interval.
				time.Sleep(time.Duration(params.Interval) * time.Millisecond)
			}

			ctx, cancel := context.WithTimeout(
				ctx, time.Duration(params.Timeout)*time.Millisecond,
			)
			defer cancel()

			// Do the pinging.
			var u *float64
			res, err := p.Ping(ctx, addr, 0)
			if err == nil {
				x := float64(res.Duration.Microseconds()) / 1000
				u = &x
			} else {
				log.Error("failed to ping", zap.Error(err))
			}

			// Log a successful ping.
			if isJson {
				jsonResponses = append(jsonResponses, &PingResponse{
					Hostname:  hostname,
					IPAddress: addr.String(),
					Latency:   u,
				})
			} else {
				if u == nil {
					strResponses = append(strResponses,
						fmt.Sprintf("%s (ping failed)", hostnameOrIp))
				} else {
					strResponses = append(strResponses,
						fmt.Sprintf("%s (time=%.3fms)", hostnameOrIp, *u))
				}
			}
		}

		// Return the responses.
		if isJson {
			ctx.JSON(200, jsonResponses)
		} else {
			ctx.String(200, strings.Join(strResponses, "\n")+"\n")
		}
	})
}
