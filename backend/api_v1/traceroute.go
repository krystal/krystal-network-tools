package api_v1

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	pingttl "github.com/strideynet/go-ping-ttl"
	"golang.org/x/sync/errgroup"
)

type tracerouteParams struct {
	// IPv6 defines if the traceroute should be ran as IPv6.
	IPv6 bool `json:"ipv6"`

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

func traceroute(g *gin.RouterGroup, pinger pinger) {
	g.GET("/:hostnameOrIp", func(c *gin.Context) {
		// Get the hostname or IP.
		hostnameOrIp := c.Param("hostnameOrIp")

		// Defines if this is JSON.
		isJson := c.ContentType() == "application/json"

		// Parse the traceroute parameters.
		var p tracerouteParams
		if err := c.BindQuery(&p); err != nil {
			c.String(400, "unable to parse query params: %s", err.Error())
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
		lookupProtocol := "ip4"
		if p.IPv6 {
			lookupProtocol = "ip6"
		}
		addr, err := net.ResolveIPAddr(lookupProtocol, hostnameOrIp)
		if err != nil {
			c.Error(&gin.Error{
				Type: gin.ErrorTypePublic,
				Err:  errors.New("unable to parse hostname or IP"),
			})
			return
		}

		// Set the default timeout.
		if p.Timeout == 0 || p.Timeout > 5000 {
			p.Timeout = 5000
		}

		// Go through each hop.
		strResponses := []string{}
		jsonResponses := []*TraceItem{}
		for _, hop := range hops {
			// Defines if the destination was reached.
			var destinationReached uintptr

			// Set the IP address and RDNS for this hop.
			var hopIp net.Addr
			var hopRdns *string
			hopIpLock := sync.Mutex{}
			setHopIpInfo := func(ip net.Addr) {
				hopIpLock.Lock()
				defer hopIpLock.Unlock()
				if hopIp != nil {
					return
				}
				hopIp = ip
				if hosts, _ := net.LookupAddr(hopIp.String()); hosts != nil && len(hosts) > 0 {
					hopRdns = &hosts[0]
				}
			}

			// Defines the array of tries.
			tries := [3]*float64{}

			// Do our 3 tries.
			eg := errgroup.Group{}
			for try := 0; try < 3; try++ {
				tryPtr := &tries[try]
				eg.Go(func() error {
					ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.Timeout)*time.Millisecond)
					resp, err := pinger.Ping(ctx, addr, int(hop))
					cancel()
					if err == nil {
						// Set the value based on the response.
						setHopIpInfo(addr)
						f := float64(resp.Duration.Microseconds()) / 1000
						*tryPtr = &f
						atomic.StoreUintptr(&destinationReached, 1)
					} else {
						// Handle the various errors that can be thrown.
						var destUnreachErr *pingttl.DestinationUnreachableErr
						var timeExceededErr *pingttl.TimeExceededErr
						if errors.As(err, &destUnreachErr) {
							// In this event, it is likely the first hop. Most traceroute systems
							// tend to just ignore this error.
							setHopIpInfo(destUnreachErr.Peer)
							f := float64(destUnreachErr.Duration.Microseconds()) / 1000
							*tryPtr = &f
						} else if errors.As(err, &timeExceededErr) {
							// The only likely information we can get from the event is the remote
							// IP address. We should get this if needed.
							setHopIpInfo(timeExceededErr.Peer)
							f := float64(timeExceededErr.Duration.Microseconds()) / 1000
							*tryPtr = &f
						} else if !errors.Is(err, context.DeadlineExceeded) {
							// Something went wrong internally.
							return err
						}
					}
					return nil
				})
			}
			if err = eg.Wait(); err != nil {
				c.Error(err)
				return
			}

			// Add the response to the slice.
			if isJson {
				// If this is nil, that's okay, that is how we repersent a timeout.
				if hopIp != nil {
					jsonResponses = append(jsonResponses, &TraceItem{
						Pings:     tries,
						RDNS:      hopRdns,
						IPAddress: hopIp.String(),
					})
				}
			} else {
				if hopIp == nil {
					strResponses = append(strResponses, strconv.FormatUint(uint64(hop), 10)+"\t*\t*\t*\t*\t")
				} else {
					resp := hopIp.String()
					if hopRdns != nil {
						resp += " (" + *hopRdns + ")"
					}
					resp += "\t"
					for _, pi := range tries {
						if pi == nil {
							resp += "*\t"
						} else {
							resp += fmt.Sprint(*pi) + "\t"
						}
					}
					strResponses = append(strResponses, resp)
				}
			}

			// Do not carry on if the destination is reached.
			if atomic.LoadUintptr(&destinationReached) == 1 {
				break
			}
		}

		// Return either JSON or string responses.
		if isJson {
			c.JSON(200, TraceResponse{
				Traceroute:    jsonResponses,
				DestinationIP: addr.String(),
			})
		} else {
			c.String(200, strings.Join(strResponses, "\n")+"\n")
		}
	})
}
