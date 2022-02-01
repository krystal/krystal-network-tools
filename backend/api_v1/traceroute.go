package api_v1

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	traceroutego "github.com/pixelbender/go-traceroute/traceroute"
)

type tracerouteParams struct {
	// Timeout is used to define how long to wait for a response from the remote host.
	Timeout uint `form:"timeout"`

	// Hop is used to define which hop should be ran.
	Hop uint `form:"hop"`

	// TotalHops is an alternative to the hop parameter which defines the total number of hops that should be ran.
	TotalHops uint `form:"total_hops"`
}

// TraceItem is used to define an item within the traceroute slice.
type TraceItem struct {
	// Pings is used to define the latency of all pings sent.
	Pings [3]*float64 `json:"pings"`

	// RDNS is used to define the RDNS of the host if valid.
	RDNS *string `json:"rdns"`

	// IPAddress is used to define the IP address of the host.
	IPAddress string `json:"ip_address"`
}

// TraceResponse is used to define the response of the traceroute API.
type TraceResponse struct {
	// Traceroute is used to define the traceroute slice.
	Traceroute []*TraceItem `json:"traceroute"`

	// DestinationIP is used to define the destination IP address.
	DestinationIP string `json:"destination_ip"`
}

func traceroute(g *gin.RouterGroup) {
	g.GET("/:hostnameOrIp", func(ctx *gin.Context) {
		// Get the hostname or IP.
		hostnameOrIp := ctx.Param("hostnameOrIp")

		// Defines if this is JSON.
		isJson := ctx.ContentType() == "application/json"

		// Parse the traceroute parameters.
		var p tracerouteParams
		if err := ctx.BindQuery(&p); err != nil {
			ctx.String(400, "unable to parse query params: %s", err.Error())
			return
		}

		// Defines all hops that need to be ran.
		if p.TotalHops > 64 || (p.Hop == 0 && p.TotalHops == 0) {
			p.TotalHops = 64
		}
		hops := []uint{p.Hop}
		if p.TotalHops != 0 {
			hops = make([]uint, p.TotalHops)
			for i := uint(0); i < p.TotalHops; i++ {
				hops[i] = i + 1
			}
		}

		// Get the addresses.
		addrs, err := net.LookupHost(hostnameOrIp)
		if err != nil {
			if isJson {
				ctx.JSON(400, map[string]string{
					"message": "invalid hostname or IP",
				})
			} else {
				ctx.String(400, "invalid hostname or IP")
			}
			return
		}
		ipAddr := net.ParseIP(addrs[0])

		// Set the default timeout.
		if p.Timeout == 0 || p.Timeout > 1000 {
			p.Timeout = 1000
		}

		// Go through each hop.
		strResponses := []string{}
		jsonResponses := []*TraceItem{}
		var s *traceroutego.Session
	hopFor:
		for _, hop := range hops {
			breakState := (func() int {
				// Get all needed replies.
				replies := [3]*traceroutego.Reply{}
				pings := [3]*float64{}
				for i := 0; i < 3; i++ {
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
					s, err = t.NewSession(ipAddr)
					if err != nil {
						ctx.Error(fmt.Errorf("failed to start traceroute session: %v", err.Error()))
						return 0
					}
					ttl := int(hop)
					if ttl == 0 {
						ttl = 1
					}
					startTime := time.Now()
					if err = s.Ping(ttl); err != nil {
						ctx.Error(fmt.Errorf("failed to ping: %v", err.Error()))
						return 0
					}

					// Handle the timeout.
					select {
					case r := <-s.Receive():
						replies[i] = r
						p := float64(time.Now().Sub(startTime).Microseconds()) / 1000
						pings[i] = &p
					case <-time.After(time.Millisecond * time.Duration(p.Timeout)):
						// Just a timeout. Do nothing.
					}
				}

				// Get the RDNS.
				var rdns *string
				if replies[0] != nil {
					// Get the replying IP address.
					replyingIp := replies[0].IP

					// Get the RDNS.
					if hosts, _ := net.LookupAddr(replyingIp.String()); hosts != nil && len(hosts) > 0 {
						rdns = &hosts[0]
					}
				}

				// Handle string or JSON formatting.
				if isJson {
					// If this is nil, that's okay, that is how we repersent a timeout.
					if replies[0] != nil {
						jsonResponses = append(jsonResponses, &TraceItem{
							Pings:     pings,
							RDNS:      rdns,
							IPAddress: replies[0].IP.String(),
						})
					}
				} else {
					if replies[0] == nil {
						strResponses = append(strResponses, strconv.FormatUint(uint64(hop), 10)+"\t*\t*\t*\t*\t")
					} else {
						resp := replies[0].IP.String()
						if rdns != nil {
							resp += " (" + *rdns + ")"
						}
						resp += "\t"
						for _, pi := range pings {
							if pi == nil {
								resp += "*\t"
							} else {
								resp += fmt.Sprint(*pi) + "\t"
							}
						}
						strResponses = append(strResponses, resp)
					}
				}

				// Break the hop for loop if the IP is the same.
				if replies[0] != nil && ipAddr.Equal(replies[0].IP) {
					return 1
				}
				return 2
			})()

			// Decide how to handle this hop.
			switch breakState {
			case 0:
				return
			case 1:
				break hopFor
			}
		}

		// Return either JSON or string responses.
		if isJson {
			ctx.JSON(200, TraceResponse{
				Traceroute:    jsonResponses,
				DestinationIP: ipAddr.String(),
			})
		} else {
			ctx.String(200, strings.Join(strResponses, "\n"))
		}
	})
}
