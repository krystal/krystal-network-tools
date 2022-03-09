package api_v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gobeam/stringy"
	dnsTools "github.com/krystal/krystal-network-tools/backend/dns"
	godns "github.com/miekg/dns"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type inMemoryCacher struct {
	sync.RWMutex
	cache map[uint16]map[string]*godns.Msg
}

func (m *inMemoryCacher) lookup(log *zap.Logger, recordType uint16, domain, nameServer string) *godns.Msg {
	m.RLock()
	defer m.RUnlock()
	recordCache := m.cache[recordType]
	if recordCache == nil {
		return nil
	}
	res, ok := recordCache[domain+"\n"+nameServer]
	if ok {
		log.Info("request cache hit!", zap.String("domain", domain), zap.String("nameServer", nameServer))
	}
	return res
}

func (m *inMemoryCacher) write(recordType uint16, domain, nameServer string, msg *godns.Msg) {
	m.Lock()
	defer m.Unlock()
	recordCache := m.cache[recordType]
	if recordCache == nil {
		recordCache = map[string]*godns.Msg{}
		m.cache[recordType] = recordCache
	}
	recordCache[domain+"\n"+nameServer] = msg
}

type cacher interface {
	lookup(log *zap.Logger, recordType uint16, domain, nameServer string) *godns.Msg
	write(recordType uint16, domain, nameServer string, msg *godns.Msg)
}

var _ cacher = &inMemoryCacher{}

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

func godnsLookup(log *zap.Logger, cacher cacher, addr string, recordType uint16, hostname string) (*godns.Msg, error) {
	// Return from cache if we have it.
	if cacher != nil {
		cached := cacher.lookup(log, recordType, hostname, addr)
		if cached != nil {
			return cached, nil
		}
	}

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
	defer conn.Close()

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
	if err == nil {
		if cacher != nil {
			cacher.write(recordType, hostname, addr, msg)
		}
	}
	return msg, err
}

func reverseStringSlice(s []string) []string {
	cpy := make([]string, len(s))
	copy(cpy, s)
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		cpy[i], cpy[j] = cpy[j], cpy[i]
	}
	return cpy
}

func findNameserver(log *zap.Logger, cacher cacher, chunks []string, recursionCount int) (string, error) {
	// Handle the recursion limit.
	if recursionCount > 10 {
		return "", errors.New("recursion limit reached")
	}

	// Reverse the chunks.
	chunksReversed := reverseStringSlice(chunks)

	// Defines the nameserver.
	ns := dnsTools.NextRootServer()

	// Handle the root nameserver.
	if len(chunks) == 1 {
		return ns, nil
	}

	// Go through each chunk until we find the end.
	for i := 0; i < len(chunksReversed); i++ {
		// Get the first i items of the slice.
		chunksToGet := chunksReversed[:i+1]

		// Get the hostname.
		hostname := strings.Join(reverseStringSlice(chunksToGet), ".") + "."

		// Turn the nameserver into an IP address.
		ipAddr, err := net.ResolveIPAddr("ip", strings.TrimRight(ns, "."))
		if err != nil {
			log.Error("failed to resolve nameserver", zap.Error(err))
			return "", err
		}
		host := ipAddr.String() + ":53"

		// Lookup the hostname.
		msg, err := godnsLookup(log, cacher, host, godns.StringToType["NS"], hostname)
		if err != nil {
			return "", err
		}

		// If we got no NS records, return an error.
		if len(msg.Answer) == 0 {
			// Handle if it's stored in msg.Ns.
			if len(msg.Ns) > 0 {
				switch t := msg.Ns[0].(type) {
				case *godns.SOA:
					ns = t.Ns
				case *godns.NS:
					ns = t.Ns
				}
				continue
			}

			// Otherwise, throw an error.
			return "", fmt.Errorf("no NS records found for %s", hostname)
		}

		// Set the first NS record to the next name server.
		switch t := msg.Answer[0].(type) {
		case *godns.SOA:
			ns = t.Ns
		case *godns.NS:
			ns = t.Ns
		case *godns.CNAME:
			// Recursively hunt for the NS record.
			return findNameserver(log, cacher, chunkifyHost(t.Target), recursionCount+1)
		default:
			return "", fmt.Errorf("unexpected record type %T", t)
		}
	}

	// Return the last name server.
	return ns, nil
}

func chunkifyHost(hostname string) []string {
	hostname = strings.TrimSuffix(hostname, ".")
	s := strings.Split(hostname, ".")
	noBlanks := make([]string, 0, len(s))
	for _, v := range s {
		if v != "" {
			noBlanks = append(noBlanks, v)
		}
	}
	return noBlanks
}

func resolveDnsLookup(log *zap.Logger, cacher cacher, nameserver, recordType, lookup string, cache bool) (responses []*DNSResponse, err error) {
	// Do the main DNS lookup.
	addr := nameserver + ":53"
	result, err := godnsLookup(log, cacher, addr, godns.StringToType[recordType], lookup)
	if err != nil {
		log.Error("failed to lookup DNS record", zap.Error(err))
		return nil, err
	}

	// Defines a higher level slice for the responses.
	responses = make([]*DNSResponse, 0, len(result.Answer))

	// Defines a map to deduplicate the responses.
	dedupe := map[string]struct{}{}

	// Go through each answer and check if we need to do any traversals.
	answers := result.Answer
	if len(answers) == 0 && recordType == "NS" {
		answers = result.Ns
	}
	for _, v := range answers {
		// Defines this records nameserver.
		answerNameserver := nameserver

		// Defines the value responses.
		valueResponses := []godns.RR{v}

		// If this is a CNAME and we do not expect this, traverse through until we find what the user is after.
		if recordType != "CNAME" {
			if startCname, ok := v.(*godns.CNAME); ok {
				recursionCount := 0
				for {
					// Check the recursion count.
					if recursionCount > 10 {
						return nil, fmt.Errorf("recursion limit reached for CNAME %s", startCname)
					}

					// Get the CNAME.
					cname := v.(*godns.CNAME)

					// Turn this host into chunks.
					chunks := chunkifyHost(cname.Target)

					// Get the nameserver if this isn't a cache.
					if !cache {
						answerNameserver, err = findNameserver(log, cacher, chunks, 0)
						if err != nil {
							return nil, err
						}
					}

					// Turn the name server into an IP address.
					ipAddr, err := net.ResolveIPAddr("ip", strings.TrimRight(answerNameserver, "."))
					if err != nil {
						log.Error("failed to resolve nameserver", zap.Error(err))
						return nil, err
					}
					host := ipAddr.String() + ":53"

					// Do the record lookup.
					traverseResult, err := godnsLookup(log, cacher, host, godns.StringToType[recordType], cname.Target)
					if err != nil {
						log.Error("failed to lookup DNS record", zap.Error(err))
						return nil, err
					}

					// If there's nothing on this host, break here.
					if len(traverseResult.Answer) == 0 {
						if recordType == "NS" {
							// Handle if it's stored in the slice.
							results := make([]godns.RR, 0, len(traverseResult.Ns))
							for _, v := range traverseResult.Ns {
								results = append(results, v)
							}
						} else {
							// There's no way there can be any records here.
							valueResponses = []godns.RR{}
						}
						break
					}

					// Handle CNAME's.
					results := make([]godns.RR, 0, len(traverseResult.Answer))
					for _, v := range traverseResult.Answer {
						if cname, ok = v.(*godns.CNAME); !ok {
							results = append(results, v)
						}
					}

					// If we have found the non-CNAME records, break here.
					if len(results) > 0 {
						valueResponses = results
						break
					}

					// Set v to the new CNAME.
					v = cname

					// Add 1 to the recursion count.
					recursionCount++
				}
			}
		}

		// Go through each value.
		for _, v := range valueResponses {
			// Get the data from the record.
			// Due to the nature of the library, this is sadly a little magical.
			var data []byte
			reflectValue := reflect.Indirect(reflect.ValueOf(v))
			reflectType := reflectValue.Type()
			n := reflectType.NumField()
			if recordType == "CNAME" {
				data, _ = json.Marshal(v.(*godns.CNAME).Target)
			} else {
				for i := 0; i < n; i++ {
					f := reflectType.Field(i)
					if strings.ToUpper(f.Name) == recordType {
						// This is the field we want.
						var err error
						data, err = json.Marshal(reflectValue.FieldByName(f.Name).Interface())
						if err != nil {
							return nil, fmt.Errorf("failed to marshal json: %v", err)
						}
						break
					}
				}
			}
			if data == nil {
				// In this situation, we will throw it into the JSON cleanifier.
				data, err = json.Marshal(jsonCleanifier{
					Value:      v,
					RemoveKeys: []string{"Hdr"},
				})
				if err != nil {
					return nil, fmt.Errorf("failed to marshal json: %v", err)
				}
			}

			// Handle the priority for MX records.
			var preference *uint16
			if mx, ok := v.(*godns.MX); ok {
				preference = &mx.Preference
			}

			// Append the result if it's unique.
			h := v.Header()
			x := &DNSResponse{
				Type:         recordType,
				TTL:          h.Ttl,
				Name:         strings.TrimRight(lookup, "."),
				Value:        data,
				Preference:   preference,
				dnsStringify: v.String,
			}
			b, _ := json.Marshal(x)
			_, ok := dedupe[string(b)]
			if !ok {
				dedupe[string(b)] = struct{}{}
				x.DNSServer = strings.TrimRight(answerNameserver, ".")
				responses = append(responses, x)
			}
		}
	}

	// No errors!
	return
}

func findAuthoritativeNameserver(log *zap.Logger, hostname string) (string, []*DNSResponse, error) {
	// Select a root nameserver to begin our search
	rootNameserver := dnsTools.NextRootServer()

	resp := []*DNSResponse{}
	var recursiveSearch func(iteration int, nameserver string) (string, error)
	recursiveSearch = func(iteration int, nameserver string) (string, error) {
		msg, err := godnsLookup(log, nil, nameserver+":53", godns.TypeNS, hostname)
		if err != nil {
			return "", err
		}

		for _, answer := range append(msg.Answer, msg.Ns...) {
			nsRecord, ok := answer.(*godns.NS)
			if !ok {
				return "", fmt.Errorf("unexpected godns type: %T", answer)
			}
			b, _ := json.Marshal(nsRecord.Ns)
			resp = append(resp, &DNSResponse{
				DNSServer:    nameserver,
				Type:         "NS",
				Name:         strings.TrimRight(nsRecord.Header().Name, "."),
				Value:        b,
				TTL:          answer.Header().Ttl,
				dnsStringify: nsRecord.String,
			})
		}

		if len(msg.Answer) > 0 {
			authoritativeNameserver := msg.Answer[rand.Intn(len(msg.Answer))]

			nameserverRecord, ok := authoritativeNameserver.(*godns.NS)
			if !ok {
				return "", fmt.Errorf("unexpected godns type: %T", authoritativeNameserver)
			}

			return strings.TrimRight(nameserverRecord.Ns, "."), nil
		}

		if len(msg.Ns) == 0 {
			// If it doesn't return a answer, or somewhere we can go for an answer
			// we are effectively lost !
			return "", errors.New("no answer or authoritive server provided in dns response")
		}

		nextNameserver := msg.Ns[rand.Intn(len(msg.Ns))]

		nameserverRecord, ok := nextNameserver.(*godns.NS)
		if !ok {
			return "", fmt.Errorf("unexpected godns type: %T", nextNameserver)
		}

		iteration += 1

		return recursiveSearch(iteration, nameserverRecord.Ns)
	}

	authoritativeNameserver, err := recursiveSearch(0, rootNameserver)
	if err != nil {
		return "", nil, err
	}

	return authoritativeNameserver, resp, nil
}

func doDnsLookups(log *zap.Logger, dnsServer, recordType, hostname string, fullTrace bool) (map[string][]*DNSResponse, error) {
	// Handle if the hostname doesn't end with a dot.
	if !strings.HasSuffix(hostname, ".") {
		hostname += "."
	}

	// Create the response map.
	responses := map[string][]*DNSResponse{}

	// Defines the mutex for the responses.
	responsesLock := sync.Mutex{}

	// Create the cacher.
	cacher := &inMemoryCacher{cache: map[uint16]map[string]*godns.Msg{}}

	// Handle non-full trace lookups.
	if !fullTrace {
		// Get the record types.
		recordTypes := []string{strings.ToUpper(recordType)}
		if recordType == "ANY" {
			// Not many DNS resolvers support this anymore, set it to literally all record types.
			recordTypes = []string{"A", "AAAA", "CNAME", "MX", "PTR", "SOA", "TXT", "NS"}
		}

		// Create the error groups.
		eg := errgroup.Group{}

		// Go through each record type and do the lookups.
		for _, recordLoop := range recordTypes {
			record := recordLoop
			eg.Go(func() error {
				r, err := resolveDnsLookup(log, cacher, dnsServer, record, hostname, true)
				if err != nil {
					return err
				}
				responsesLock.Lock()
				responses[record] = r
				responsesLock.Unlock()
				return nil
			})
		}

		// Wait for the group to finish and then return the results.
		if err := eg.Wait(); err != nil {
			return nil, err
		}
		return responses, nil
	}

	// Handle a full trace search
	authoritativeNameserver, answer, err := findAuthoritativeNameserver(log, hostname)
	if err != nil {
		return nil, err
	}

	responses["NS"] = answer

	// Get the record types.
	recordTypes := []string{strings.ToUpper(recordType)}
	if recordType == "ANY" {
		// Not many DNS resolvers support this anymore, set it to literally all record types.
		recordTypes = []string{"A", "AAAA", "CNAME", "MX", "PTR", "SOA", "TXT", "NS"}
	} else if recordTypes[0] == "NS" {
		// We already have this data. Make this a blank slice.
		recordTypes = []string{}
	}

	eg := errgroup.Group{}
	// Spawn a goroutine to look up each record type.
	for _, recordLoop := range recordTypes {
		record := recordLoop
		eg.Go(func() error {
			r, err := resolveDnsLookup(log, cacher, authoritativeNameserver, record, hostname, false)
			if err != nil {
				return err
			}
			responsesLock.Lock()
			x := responses[record]
			if x == nil {
				x = []*DNSResponse{}
			}
			responses[record] = append(x, r...)
			responsesLock.Unlock()
			return nil
		})
	}

	// Wait for the results.
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Return the responses.
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
		if hostname == "" {
			context.Error(&gin.Error{
				Type: gin.ErrorTypePublic,
				Err:  errors.New("invalid hostname"),
			})
			return
		}

		// Do the DNS lookup.
		results, err := doDnsLookups(log, dnsServer, recordType, hostname, params.Trace)
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
