package api_v1

import (
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	goping "github.com/go-ping/ping"
)

type pingParams struct {
	Timeout  uint `form:"timeout"`
	Count    uint `form:"count"`
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

	// Hostname is used to define the hostname. If it is blank, it means the rDNS lookup is not available.
	Hostname string `json:"hostname,omitempty"`

	// Latency is used to define the latency.
	Latency *uint64 `json:"latency,omitempty"`
}

func ping(g *gin.RouterGroup) {
	g.GET("/:hostnameOrIp", func(ctx *gin.Context) {
		// Get the hostname or IP.
		hostnameOrIp := ctx.Param("hostnameOrIp")

		// Defines if this is JSON.
		isJson := ctx.ContentType() == "application/json"

		// Bind the ping params.
		var p pingParams
		if err := ctx.BindQuery(&p); err != nil {
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
		if p.Count > 0 {
			if p.Count > 10 {
				p.Count = 10
			}
		} else {
			p.Count = 1
		}

		// Attempt to make the initial pinger to get the IP address.
		pinger, err := goping.NewPinger(hostnameOrIp)
		if err != nil {
			if isJson {
				ctx.JSON(400, map[string]string{
					"message": err.Error(),
				})
			} else {
				ctx.String(400, "unable to resolve %s: %s", hostnameOrIp, err.Error())
			}
			return
		}
		addr := pinger.IPAddr()

		// Attempt a rdns lookup.
		hostnameOrIp = addr.String()
		hostname := ""
		if hosts, _ := net.LookupAddr(hostnameOrIp); hosts != nil && len(hosts) > 0 {
			hostnameOrIp += " [" + hosts[0] + "]"
			hostname = hosts[0]
		}

		// Defines all responses.
		strResponses := []string{}
		jsonResponses := []*PingResponse{}

		// Make sure the interval is less than or equal to 1 second.
		if p.Interval > 1000 {
			p.Interval = 1000
		}

		// Run each ping within its own context to make sure that dropped packets are logged in order.
		for i := uint(0); i < p.Count; i++ {
			// If i isn't 0, sleep for the specified interval and make a new pinger with the address.
			if i != 0 {
				// Sleep for the specified interval.
				time.Sleep(time.Duration(p.Interval) * time.Millisecond)

				// Make the new pinger.
				pinger, err = goping.NewPinger(addr.String())
				if err != nil {
					if isJson {
						ctx.JSON(400, map[string]string{
							"message": err.Error(),
						})
					} else {
						ctx.String(400, "unable to resolve %s: %s", hostnameOrIp, err.Error())
					}
					return
				}
			}

			// This is a special hack because the pinger will wait a second before starting if we do not set this.
			pinger.Interval = time.Duration(1)

			// Set the timeout duration. Note we do not use the go-ping timeout function. This is because
			// it will block for the timeout duration.
			d := time.Second
			if p.Timeout > 0 {
				if p.Timeout > 1000 {
					p.Timeout = 1000
				}
				d = time.Duration(p.Timeout) * time.Millisecond
			}

			// Set the ping count.
			pinger.Count = 1
			if p.Count > 0 {
				if p.Count > 10 {
					p.Count = 10
				}
				pinger.Count = int(p.Count)
			}

			// Make sure that we block until the first packet is received, or we time out.
			var singleSend uintptr
			errChan := make(chan error, 1)
			flushChan := func(err error) {
				if atomic.SwapUintptr(&singleSend, 1) == 1 {
					return
				}
				errChan <- err
				close(errChan)
			}
			t := time.AfterFunc(d, func() {
				pinger.Stop()
				flushChan(nil)
			})
			go func() {
				// Call the run function.
				innerErr := pinger.Run()

				// Stop the timer if it hasn't gone off.
				t.Stop()

				// Flush the channel.
				flushChan(innerErr)
			}()

			// Wait for the channel.
			if err = <-errChan; err != nil {
				if isJson {
					ctx.JSON(400, map[string]string{
						"message": err.Error(),
					})
				} else {
					ctx.String(400, "unable to ping %s: %s", hostnameOrIp, err.Error())
				}
				return
			}

			// Log timeouts.
			s := pinger.Statistics()
			if len(s.Rtts) == 0 {
				if isJson {
					jsonResponses = append(jsonResponses, &PingResponse{
						Error: &PingErrorMessage{
							IsTimeout: true,
							Message:   "response timeout",
						},
						Hostname: hostname,
					})
				} else {
					strResponses = append(strResponses,
						fmt.Sprintf("unable to ping %s: response timeout", hostnameOrIp))
				}
				continue
			}

			// Log a successful ping.
			if isJson {
				u := uint64(s.Rtts[0].Milliseconds())
				jsonResponses = append(jsonResponses, &PingResponse{
					Hostname: hostname,
					Latency:  &u,
				})
			} else {
				strResponses = append(strResponses,
					fmt.Sprintf("%d bytes from %s (time=%dms)", pinger.Size, hostnameOrIp,
						s.Rtts[0].Milliseconds()))
			}
		}

		// Return the responses.
		if isJson {
			ctx.JSON(200, jsonResponses)
		} else {
			ctx.String(200, strings.Join(strResponses, "\n"))
		}
	})
}
