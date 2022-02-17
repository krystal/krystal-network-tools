package api_v1

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	traceroutego "github.com/jakemakesstuff/traceroute"
)

var (
	currentPort     uint = 34457
	currentPortLock      = sync.Mutex{}
)

func getPort() uint {
	currentPortLock.Lock()
	defer currentPortLock.Unlock()
	if currentPort > 40000 {
		currentPort = 34457
	}
	currentPort++
	return currentPort
}

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

		// Get the local interfaces.
		ifaceAddrs, err := net.InterfaceAddrs()
		if err != nil {
			ctx.Error(err)
			return
		}

		// Find the relevant local IP.
		var localIP net.IP
		is6 := ipAddr.To4() == nil
		if is6 {
			// Look for a IPv6 address.
			for _, ifaceAddr := range ifaceAddrs {
				if ipnet, ok := ifaceAddr.(*net.IPNet); ok && ipnet.IP.To4() == nil && !ipnet.IP.IsLoopback() {
					localIP = ipnet.IP
					break
				}
			}
		} else {
			// Look for a IPv4 address.
			for _, ifaceAddr := range ifaceAddrs {
				if ipnet, ok := ifaceAddr.(*net.IPNet); ok && ipnet.IP.To4() != nil && !ipnet.IP.IsLoopback() {
					localIP = ipnet.IP
					break
				}
			}
		}

		// Go through each hop.
		strResponses := []string{}
		jsonResponses := []*TraceItem{}
	hopFor:
		for _, hop := range hops {
			breakState := (func() int {
				// Get all needed replies.
				pings := [3]*float64{}

				if is6 {
					// Add 1 to the hop. The first hop will without fail always not work.
					hop++
				}

				port := int(getPort())
				for i := 0; i < 2; i++ {
					// In this library, error is used solely to signify when the TTL has been
					// sent before the request is done. We can very safely ignore this error.
					// We may want to re-evaluate this in future releases of the library, but
					// since we are likely just going to pin this, it does not make a huge difference.
					res, _ := traceroutego.Traceroute(&traceroutego.TracerouteOptions{
						SourcePort:      port,
						SourceAddr:      localIP,
						ProbeType:       traceroutego.IcmpProbe,
						DestinationAddr: ipAddr,
						DestinationPort: 33434,
						StartingTTL:     int(hop),
						MaxTTL:          int(hop),
						ProbeCount:      3,
						ProbeTimeout:    time.Duration(p.Timeout) * time.Millisecond,
					})

					// Check the hop is there.
					if len(res.Hops) != 1 {
						ctx.Error(fmt.Errorf("probe count wrong"))
						return 0
					}
					hopObj := res.Hops[0]

					// Get the RDNS.
					var rdns *string
					var firstNonNull *traceroutego.ProbeResponse
					for _, v := range hopObj.Responses {
						if v.Success {
							// Get the replying IP address.
							replyingIp := v.Address

							// Get the RDNS.
							if hosts, _ := net.LookupAddr(replyingIp.String()); hosts != nil && len(hosts) > 0 {
								rdns = &hosts[0]
							}

							// Set the value and break.
							firstNonNull = &v
							break
						}
					}

					// Handle adding all the pings.
					for i, v := range hopObj.Responses {
						if v.Success {
							f := float64(v.Duration.Microseconds()) / 1000
							pings[i] = &f
						}
					}

					// Handle string or JSON formatting.
					if isJson {
						// If this is nil, that's okay, that is how we repersent a timeout.
						if firstNonNull != nil {
							jsonResponses = append(jsonResponses, &TraceItem{
								Pings:     pings,
								RDNS:      rdns,
								IPAddress: firstNonNull.Address.String(),
							})
						}
					} else {
						if firstNonNull == nil {
							strResponses = append(strResponses, strconv.FormatUint(uint64(hop), 10)+"\t*\t*\t*\t*\t")
						} else {
							resp := firstNonNull.Address.String()
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
					if firstNonNull != nil {
						if ipAddr.Equal(firstNonNull.Address) {
							return 1
						}
						break
					}
				}

				// Process the next hop.
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
