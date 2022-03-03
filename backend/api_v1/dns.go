package api_v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gobeam/stringy"
	godns "github.com/miekg/dns"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type dnsParams struct {
	// Trace is used to define if the DNS record should be traced all the way to the nameserver.
	Trace bool `form:"trace"`
}

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

// DNSResponse is used to define the JSON responses for the DNS API.
type DNSResponse struct {
	// Type is used to define the type of the record.
	Type string `json:"type"`

	// TTL is the time to live of the DNS record.
	TTL uint32 `json:"ttl"`

	// Preference is used for MX records.
	Preference *uint16 `json:"priority,omitempty"`

	// Name is used to define the name of the DNS record.
	Name string `json:"name"`

	// DNSServer defines the DNS server which gave this response.
	DNSServer string `json:"dnsServer"`

	// Value is used to define the value of the DNS record.
	Value json.RawMessage `json:"value"`

	// Defines the DNS string.
	dnsStringify func() string
}

func godnsLookup(log *zap.Logger, addr string, recordType uint16, hostname string) (*godns.Msg, error) {
	// Create the DNS message.
	msg := &godns.Msg{}
	msg.Id = godns.Id()
	msg.RecursionDesired = true

	// DNS servers prefer 1 message per request. Make the question.
	msg.Question = []godns.Question{{
		Name:   hostname,
		Qtype:  recordType,
		Qclass: godns.StringToClass["IN"],
	}}
	conn, err := godns.Dial("tcp", addr)
	if err != nil {
		log.Error("failed to connect to dns server", zap.Error(err))
		return nil, err
	}

	// Send the DNS message.
	err = conn.WriteMsg(msg)
	if err != nil {
		return nil, err
	}

	// Read the DNS response.
	msg, err = conn.ReadMsg()
	if err != nil {
		log.Error("failed to read from dns server", zap.Error(err))
	}
	return msg, err
}

func findNameserverHostname(log *zap.Logger, addr string, chunks []string) (string, int, uint32, error) {
	var msg *godns.Msg
	var err error
	for i := 0; i < len(chunks); i++ {
		// Compile this set of chunks.
		hostname := strings.Join(chunks[i:], ".") + "."

		// Do the DNS lookup.
		recursionCount := 0
	lookup:
		msg, err = godnsLookup(log, addr, godns.StringToType["NS"], hostname)
		if err != nil {
			continue
		}

		// Find the answer.
		if len(msg.Answer) > 0 {
			switch x := msg.Answer[0].(type) {
			case *godns.NS:
				return x.Ns, i, x.Hdr.Ttl, nil
			case *godns.CNAME:
				if recursionCount == 50 {
					return "", 0, 0, fmt.Errorf("recursion limit reached on %s", hostname)
				}
				hostname = x.Target
				recursionCount++
				goto lookup
			default:
				return "", 0, 0, errors.New("invalid type for NS record")
			}
		}
	}
	if err != nil {
		// Errored whilst trying to find NS record.
		return "", 0, 0, err
	}

	// Unable to find NS record.
	log.Warn("unable to find NS record", zap.String("hostname", strings.Join(chunks, ".")))
	return "", 0, 0, nil
}

func doDnsLookups(log *zap.Logger, dnsServer, recordType string, recursive bool, chunks []string) (map[string][]*DNSResponse, error) {
	// Resolve the IP of the DNS server.
	var addr string
	server2addr := func() error {
		rawAddr, err := net.ResolveIPAddr("ip", dnsServer)
		if err != nil {
			return err
		}
		addr = rawAddr.IP.String() + ":53"
		return nil
	}
	if err := server2addr(); err != nil {
		return nil, err
	}

	// Keep going through chunks until we get a NS record.
	initAddr := addr
	oldDnsServer := dnsServer
	host, i, ttl, err := findNameserverHostname(log, addr, chunks)
	if err != nil {
		return nil, err
	}
	if host == "" {
		// Unable to find NS record.
		log.Warn("unable to find NS record", zap.String("hostname", strings.Join(chunks, ".")))
	} else {
		// Turn it into the address.
		dnsServer = host
		if err = server2addr(); err != nil {
			return nil, err
		}
	}

	// If this was a NS lookup that is non-recursive, we have our result here.
	if !recursive && recordType == "NS" {
		if i == 0 {
			// This means that the NS record was on the record specified.
			return map[string][]*DNSResponse{
				"NS": {
					{
						Type:      recordType,
						TTL:       ttl,
						Name:      strings.TrimRight(dnsServer, "."),
						DNSServer: oldDnsServer,
					},
				},
			}, nil
		}

		// No DNS records were found.
		return map[string][]*DNSResponse{"NS": {}}, nil
	}

	// Try to update the DNS server used here.
	if err = server2addr(); err != nil {
		return nil, err
	}

	// Define the record types.
	recordTypes := []string{recordType}
	if recordType == "ANY" {
		// Not many DNS resolvers support this anymore, set it to literally all record types.
		recordTypes = []string{"A", "AAAA", "CNAME", "MX", "NS", "PTR", "SOA", "TXT"}
	}
	recordTypesPacket := make([]uint16, len(recordTypes))
	for i, v := range recordTypes {
		packetType, ok := godns.StringToType[v]
		if !ok {
			return nil, errors.New("invalid record type: " + v)
		}
		recordTypesPacket[i] = packetType
	}

	// Defines all DNS responses.
	responses := map[string][]*DNSResponse{}
	responsesLock := sync.Mutex{}
	appendToRecordType := func(recordType string, responseArgs ...*DNSResponse) {
		responsesLock.Lock()
		defer responsesLock.Unlock()

		// Stops an append with zero items being nil.
		a := responses[recordType]
		if a == nil {
			a = []*DNSResponse{}
		}

		responses[recordType] = append(a, responseArgs...)
	}
	eg := errgroup.Group{}
	for i, recordLoop := range recordTypes {
		// Get all items which may not be thread safe.
		record := recordLoop
		packetType := recordTypesPacket[i]

		// Do the DNS lookup.
		eg.Go(func() error {
			// Do the DNS lookup.
			msg, err := godnsLookup(log, addr, packetType, strings.Join(chunks, ".")+".")
			if err != nil {
				return err
			}

			// Make each response.
			dnsResponses := make([]*DNSResponse, 0)
		answerIteration:
			for _, v := range msg.Answer {
				// Handle the various responses.
				var data json.RawMessage
				originalValue := v
				resultDnsHost := dnsServer
			parseAnswer:
				switch x := v.(type) {
				case *godns.CNAME:
					if record == "CNAME" {
						// This is to be expected here since we are looking for CNAME records.
						b, _ := json.Marshal(x.Target)
						data = b
					} else {
						// In this situation, the DNS configuration is telling us to look elsewhere.
						recursionCount := 0
						for recursionCount < 50 {
							// Chunkify the CNAME.
							chunkifyReady := x.Target
							if strings.HasSuffix(chunkifyReady, ".") {
								chunkifyReady = strings.TrimRight(chunkifyReady, ".")
							}
							cnameChunks := strings.Split(chunkifyReady, ".")

							// Get the NS host.
							nsHost, _, _, err := findNameserverHostname(log, initAddr, cnameChunks)
							if err != nil {
								return err
							}
							if nsHost == "" {
								// Unable to find NS record.
								log.Warn("unable to find NS record", zap.String("hostname", strings.Join(cnameChunks, ".")))
								continue answerIteration
							}

							// Turn that into the address.
							rawAddr, err := net.ResolveIPAddr("ip", nsHost)
							if err != nil {
								return err
							}
							addr := rawAddr.IP.String() + ":53"

							// Lookup the CNAME's value.
							msg, err = godnsLookup(log, addr, packetType, x.Target)
							if err != nil {
								return err
							}

							// If there is no answers, continue the root loop.
							if len(msg.Answer) == 0 {
								continue answerIteration
							}

							// Check if this contains non-CNAME records.
							for _, iface := range msg.Answer {
								switch x := iface.(type) {
								case *godns.CNAME:
									// Ignore this.
								default:
									// We are past CNAME's!
									v = x
									resultDnsHost = nsHost
									goto parseAnswer
								}
							}

							// Set the next CNAME we are parsing.
							x = msg.Answer[0].(*godns.CNAME)

							// Add 1 to the recursion count.
							recursionCount++
						}
						return fmt.Errorf("record type %s for host %s has hit recursion limit", record, strings.Join(chunks, "."))
					}
				default:
					// Get the data from the record.
					// Due to the nature of the library, this is sadly a little magical.
					reflectValue := reflect.Indirect(reflect.ValueOf(v))
					reflectType := reflectValue.Type()
					n := reflectType.NumField()
					for i := 0; i < n; i++ {
						f := reflectType.Field(i)
						if strings.ToUpper(f.Name) == record {
							// This is the field we want.
							var err error
							data, err = json.Marshal(reflectValue.FieldByName(f.Name).Interface())
							if err != nil {
								return fmt.Errorf("failed to marshal json: %v", err)
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
							return fmt.Errorf("failed to marshal json: %v", err)
						}
					}
				}

				// Handle the priority for MX records.
				var preference *uint16
				if mx, ok := v.(*godns.MX); ok {
					preference = &mx.Preference
				}

				// Make the response.
				h := originalValue.Header()
				r := &DNSResponse{
					Type:         recordType,
					TTL:          h.Ttl,
					Name:         strings.TrimRight(h.Name, "."),
					Value:        data,
					Preference:   preference,
					DNSServer:    strings.TrimRight(resultDnsHost, "."),
					dnsStringify: v.String,
				}
				dnsResponses = append(dnsResponses, r)
			}
			appendToRecordType(record, dnsResponses...)
			return nil
		})
	}

	// Handle any additional recursion.
	mapChunks := []map[string][]*DNSResponse{}
	if recursive {
		mapChunks = make([]map[string][]*DNSResponse, len(chunks)-1)
		for i = 1; i < len(chunks); i++ {
			mapPtr := &mapChunks[i-1]
			x := i
			eg.Go(func() error {
				remainderChunks := chunks[x:]
				map_, err := doDnsLookups(log, oldDnsServer, recordType, false, remainderChunks)
				if err != nil {
					return err
				}
				*mapPtr = map_
				return nil
			})
		}
	}

	// Go ahead and run the DNS lookups.
	if err = eg.Wait(); err != nil {
		return nil, err
	}

	// Add all the map keys found in the right order and later.
	for _, map_ := range mapChunks {
		for k, v := range map_ {
			responses[k] = append(responses[k], v...)
		}
	}

	// Return all responses.
	return responses, nil
}

func dns(g *gin.RouterGroup, log *zap.Logger, dnsServer string) {
	g.GET("/:recordType/:hostname", func(context *gin.Context) {
		// Defines if this is JSON.
		isJson := context.ContentType() == "application/json"

		// Bind the params.
		var params dnsParams
		if err := context.BindQuery(&params); err != nil {
			if isJson {
				context.JSON(400, map[string]string{
					"message": err.Error(),
				})
			} else {
				context.String(400, "unable to parse query params: %s", err.Error())
			}
			return
		}

		// Get the type and hostname from the URL.
		recordType := context.Param("recordType")
		hostname := strings.TrimSuffix(context.Param("hostname"), ".")
		chunks := []string{}
		for _, v := range strings.Split(hostname, ".") {
			if v != "" {
				chunks = append(chunks, v)
			}
		}
		if len(chunks) == 0 {
			context.Error(&gin.Error{
				Type: gin.ErrorTypePublic,
				Err:  errors.New("invalid hostname"),
			})
			return
		}

		// Do the DNS lookup.
		results, err := doDnsLookups(log, dnsServer, recordType, params.Trace, chunks)
		if err != nil {
			context.Error(&gin.Error{
				Type: gin.ErrorTypePublic,
				Err:  fmt.Errorf("failed to perform dns lookup: %v", err),
			})
			return
		}

		// Handle JSON responses.
		if isJson {
			context.JSON(200, results)
			return
		}

		// Get the keys and order them.
		keys := make([]string, len(results))
		i := 0
		for k := range results {
			keys[i] = k
			i++
		}
		sort.Strings(keys)

		// Formulate the text response.
		strResponse := ""
		for _, key := range keys {
			// Get the slice.
			s := results[key]

			// Go through each value.
			for _, value := range s {
				split := strings.SplitN(value.dnsStringify(), "\t", 4)
				if strResponse != "" {
					strResponse += "\n"
				}
				strResponse += split[0] + "\t" + split[3]
			}
		}
		context.String(200, strResponse)
	})
}
