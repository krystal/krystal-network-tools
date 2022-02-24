package api_v1

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	goping "github.com/go-ping/ping"
	godns "github.com/miekg/dns"
	"go.uber.org/zap"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	currentPort     uint = 34457
	currentPortLock      = sync.Mutex{}
)

func getPort() uint {
	currentPortLock.Lock()
	defer currentPortLock.Unlock()
	if currentPort > 33523 {
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

func traceroute(g *gin.RouterGroup, logger *zap.Logger, cachedDnsServer string) {
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
		ipAddr := net.ParseIP(hostnameOrIp)
		if ipAddr == nil {
			// Make the DNS connection.
			dnsServer := cachedDnsServer
			conn, err := godns.Dial("tcp", dnsServer)
			if err != nil {
				logger.Error("failed to connect to dns server", zap.Error(err))
				ctx.Error(err)
				return
			}

			// Defer killing the connection to stop leaks.
			defer conn.Close()

			// Create the DNS message.
			msg := &godns.Msg{}
			msg.Id = godns.Id()
			msg.RecursionDesired = true

			// Make the A request.
			if !strings.HasSuffix(hostnameOrIp, ".") {
				hostnameOrIp += "."
			}
			msg.Question = []godns.Question{{
				Name:   hostnameOrIp,
				Qtype:  godns.StringToType["A"],
				Qclass: godns.StringToClass["IN"],
			}}

			// Send the DNS message.
			err = conn.WriteMsg(msg)
			if err != nil {
				ctx.Error(&gin.Error{
					Err:  fmt.Errorf("failed to perform dns lookup: %v", err),
					Type: gin.ErrorTypePublic,
				})
				return
			}

			// Read the DNS response.
			msg, err = conn.ReadMsg()
			if err != nil {
				ctx.Error(err)
				return
			}

			// Handle both cases here.
			if msg.Answer == nil || len(msg.Answer) == 0 {
				// Remake the message for an AAAA record.
				msg = &godns.Msg{}
				msg.Id = godns.Id()
				msg.RecursionDesired = true
				msg.Question = []godns.Question{{
					Name:   hostnameOrIp,
					Qtype:  godns.StringToType["AAAA"],
					Qclass: godns.StringToClass["IN"],
				}}
				err = conn.WriteMsg(msg)
				if err != nil {
					ctx.Error(&gin.Error{
						Err:  fmt.Errorf("failed to perform dns lookup: %v", err),
						Type: gin.ErrorTypePublic,
					})
					return
				}
				msg, err = conn.ReadMsg()
				if err != nil {
					ctx.Error(err)
					return
				}
				if msg.Answer != nil && len(msg.Answer) > 0 {
					ipAddr = msg.Answer[0].(*godns.AAAA).AAAA
				}
			} else {
				// Get the A record.
				aRecord := msg.Answer[0].(*godns.A)
				ipAddr = aRecord.A
			}

			// In this situation, attempt to parse the host.
			if ipAddr == nil {
				if isJson {
					ctx.JSON(400, map[string]string{
						"message": "invalid hostname or IP",
					})
				} else {
					ctx.String(400, "invalid hostname or IP")
				}
				return
			}
		}

		// Set the default timeout.
		if p.Timeout == 0 || p.Timeout > 1000 {
			p.Timeout = 1000
		}

		// Go through each hop.
		strResponses := []string{}
		jsonResponses := []*TraceItem{}
		for _, hop := range hops {
			//items := [3]*float64{}
			for i := 0; i < 3; i++ {
				// Make the pinger.
				pinger, err := goping.NewPinger(ipAddr.String())
				if err != nil {
					ctx.Error(err)
					return
				}
				pinger.TTL = int(hop)
				pinger.Interval = time.Second
				pinger.Count = 1

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
				t := time.AfterFunc(time.Second, func() {
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

				// Handle any errors.
				if err = <-errChan; err != nil {
					ctx.Error(err)
					return
				}

				stats := pinger.Statistics()
				b, _ := json.Marshal(stats)
				fmt.Println(string(b))
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
