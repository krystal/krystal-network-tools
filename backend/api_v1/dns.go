package api_v1

import (
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gobeam/stringy"
	"github.com/krystal/krystal-network-tools/backend/utils"
	godns "github.com/miekg/dns"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Used to clean the case of things in a value for JSON and remove unwanted keys.
type jsonCleanifier struct {
	// Value is used to define the JSON value.
	Value interface{}

	// RemoveKeys is keys that should be removed from the JSON value.
	RemoveKeys []string
}

// MarshalJSON implements json.Marshaler.
func (j jsonCleanifier) MarshalJSON() ([]byte, error) {
	v := reflect.Indirect(reflect.ValueOf(j.Value))
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return json.Marshal(j.Value)
	}
	n := t.NumField()
	m := map[string]json.RawMessage{}
outerFor:
	for i := 0; i < n; i++ {
		// Get the field type.
		f := t.Field(i)

		// If the field name is disallowed, ignore it.
		for _, v := range j.RemoveKeys {
			if f.Name == v {
				continue outerFor
			}
		}

		// Add this to the map.
		b, err := json.Marshal(jsonCleanifier{Value: v.Field(i).Interface()})
		if err != nil {
			return nil, err
		}
		m[stringy.New(f.Name).ToLower()] = b
	}

	// Marshal the map and return.
	return json.Marshal(m)
}

type clientErrorWrapper struct {
	error
}

// DNSResponse is used to define the JSON responses for the DNS API.
type DNSResponse struct {
	// Type is used to define the type of the record.
	Type string `json:"type"`

	// TTL is the time to live of the DNS record.
	TTL uint32 `json:"ttl"`

	// Name is used to define the name of the DNS record.
	Name string `json:"name"`

	// Value is used to define the value of the DNS record.
	Value json.RawMessage `json:"value"`
}

func dns(g *gin.RouterGroup, log *zap.Logger) {
	g.GET("/:recordType/:hostname", func(context *gin.Context) {
		// Get the type and hostname from the URL.
		recordType := context.Param("recordType")
		hostname := context.Param("hostname")
		if !strings.HasSuffix(hostname, ".") {
			hostname += "."
		}

		// Defines if this is JSON.
		isJson := context.ContentType() == "application/json"

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
		sort.Strings(recordTypes)

		// Handle formatting the results.
		strResponses := []string{}
		jsonResponses := map[string][]DNSResponse{}
		var i int
		for i, recordType = range recordTypes {
			// Get the response from the DNS server.
			response := results[i]
			if response.Answer == nil {
				// In the case that this is JSON, we don't want to return a nil array.
				if isJson {
					jsonResponses[recordType] = []DNSResponse{}
				}
			} else {
				if isJson {
					a := make([]DNSResponse, len(response.Answer))
					for i, v := range response.Answer {
						// Get the data from the record.
						// Due to the nature of the library, this is sadly a little magical.
						var data json.RawMessage
						reflectValue := reflect.Indirect(reflect.ValueOf(v))
						reflectType := reflectValue.Type()
						n := reflectType.NumField()
						for i := 0; i < n; i++ {
							f := reflectType.Field(i)
							if strings.ToUpper(f.Name) == recordType {
								// This is the field we want.
								var err error
								data, err = json.Marshal(reflectValue.FieldByName(f.Name).Interface())
								if err != nil {
									log.Error("failed to marshal json", zap.Error(err))
									context.String(500, "Internal Server Error")
									return
								}
								break
							}
						}
						if data == nil {
							// In this situation, we will throw it into the JSON cleanifier.
							var err error
							data, err = json.Marshal(jsonCleanifier{
								Value:      v,
								RemoveKeys: []string{"Hdr"},
							})
							if err != nil {
								log.Error("failed to marshal json", zap.Error(err))
								context.String(500, "Internal Server Error")
								return
							}
						}

						// Get the response.
						h := v.Header()
						a[i] = DNSResponse{
							Type:  recordType,
							TTL:   h.Ttl,
							Name:  h.Name,
							Value: data,
						}
					}
					jsonResponses[recordType] = a
				} else {
					// Use the string representation from the DNS library but remove a few chunks.
					for _, v := range response.Answer {
						s := strings.SplitN(v.String(), "\t", 4)
						strResponses = append(strResponses, s[0]+"\t"+s[3])
					}
				}
			}
		}

		// Return the response.
		if isJson {
			context.JSON(200, jsonResponses)
		} else {
			context.String(200, strings.Join(strResponses, "\n"))
		}
	})
}
