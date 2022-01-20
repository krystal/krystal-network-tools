package api_v1

import (
	"errors"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/krystal/krystal-network-tools/backend/utils"
	godns "github.com/miekg/dns"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type clientErrorWrapper struct {
	error
}

func dns(g *gin.RouterGroup, log *zap.Logger) {
	g.GET("/:recordType/:hostname", func(context *gin.Context) {
		// Get the type and hostname from the URL.
		recordType := context.Param("recordType")
		hostname := context.Param("hostname")
		if !strings.HasSuffix(hostname, ".") {
			hostname += "."
		}

		// Make the record type upper case.
		recordType = strings.ToUpper(recordType)

		// Defines the record types.
		recordTypes := []string{recordType}
		recordTypePacket, ok := godns.StringToType[recordType]
		if !ok {
			context.String(400, "Invalid record type")
			return
		}
		recordTypesPacket := []uint16{recordTypePacket}
		if recordType == "ANY" {
			// Since DNS servers rarely support ANY, we need to manually handle it.
			recordTypes = []string{"A", "AAAA", "CNAME", "MX", "NS", "PTR", "SOA", "SRV", "TXT"}
			recordTypesPacket = make([]uint16, len(recordTypes))
			for i, v := range recordTypes {
				recordTypesPacket[i], _ = godns.StringToType[v]
			}
		}

		// Defines the results.
		results := make([]*godns.Msg, len(recordTypes))

		// Go through each record to make the message.
		anyQclass := godns.StringToClass["IN"]
		wg := errgroup.Group{}
		for i, v := range recordTypesPacket {
			resultPtr := &results[i]
			qtype := v
			wg.Go(func() error {
				// Make the DNS connection.
				conn, err := godns.Dial("tcp", utils.DNSServer)
				if err != nil {
					log.Error("failed to connect to dns server", zap.Error(err))
					return err
				}

				// Defer killing the connection to stop leaks.
				defer conn.Close()

				// Create the DNS message.
				msg := &godns.Msg{}
				msg.Id = godns.Id()
				msg.RecursionDesired = true

				// DNS servers prefer 1 message per request. Make the question.
				msg.Question = []godns.Question{{
					Name:   hostname,
					Qtype:  qtype,
					Qclass: anyQclass,
				}}

				// Send the DNS message.
				err = conn.WriteMsg(msg)
				if err != nil {
					return clientErrorWrapper{err}
				}

				// Read the DNS response.
				msg, err = conn.ReadMsg()
				if err != nil {
					log.Error("failed to read from dns server", zap.Error(err))
					return err
				}

				// Set the pointer to the result and return no errors.
				*resultPtr = msg
				return nil
			})
		}

		// Handle any errors.
		if err := wg.Wait(); err != nil {
			if errors.Is(err, clientErrorWrapper{}) {
				context.String(400, "failed to perform lookup: "+err.Error())
			} else {
				context.String(500, "Internal Server Error")
			}
			return
		}

		// Sort the types by alphabetical order.
		responses := []string{}
		sort.Strings(recordTypes)
		var i int
		for i, recordType = range recordTypes {
			// Get the response from the DNS server.
			response := results[i]
			if response.Answer != nil {
				for _, v := range response.Answer {
					s := strings.SplitN(v.String(), "\t", 4)
					responses = append(responses, s[0]+"\t"+s[3])
				}
			}
		}

		// Return the response.
		context.String(200, strings.Join(responses, "\n"))
	})
}
