package dns

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"reflect"
	"strings"
	"sync"

	"github.com/gobeam/stringy"
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

type Response map[string]RecordType

func (r Response) String() string {
	str := ""
	for t, record := range r {
		str += "--- " + t + " ---\n" + record.String()
	}

	return str
}

type RecordType []Server

func (rt RecordType) String() string {
	str := ""
	for _, srv := range rt {
		str += srv.String() + "\n"
	}

	return str
}

type Server struct {
	Server  string   `json:"server"`
	Records []Record `json:"records"`
}

func (srv Server) String() string {
	str := "-- " + srv.Server + " --\n"
	for _, record := range srv.Records {
		if record.stringer != nil {
			str += record.stringer() + "\n"
		}
	}

	return str
}

type Record struct {
	// Type is used to define the type of the record.
	Type string `json:"type"`

	// TTL is the time to live of the DNS record.
	TTL uint32 `json:"ttl"`

	// Preference is used for MX records.
	Preference *uint16 `json:"priority,omitempty"`

	// Name is used to define the name of the DNS record.
	Name string `json:"name"`

	// Value is used to define the value of the DNS record.
	Value json.RawMessage `json:"value"`

	// A function returning a string version of this record.
	stringer func() string
}

// rawQuery sends a DNS request to server specified by addr.
// It returns the raw dns response.
func rawQuery(
	log *zap.Logger,
	addr string,
	recordType uint16,
	hostname string,
) (*godns.Msg, error) {
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
	return msg, err
}

func recordFromAnswer(answer godns.RR) (Record, error) {
	header := answer.Header()

	var value interface{}
	switch casted := answer.(type) {
	case *godns.A:
		value = casted.A
	case *godns.AAAA:
		value = casted.AAAA
	case *godns.CNAME:
		value = casted.Target
	case *godns.MX:
		value = casted.Mx
	case *godns.NS:
		value = casted.Ns
	case *godns.PTR:
		value = casted.Ptr
	case *godns.TXT:
		value = casted.Txt
	default:
		value = casted
	}
	data, err := json.Marshal(jsonCleanifier{
		Value:      value,
		RemoveKeys: []string{"Hdr"},
	})
	if err != nil {
		return Record{}, fmt.Errorf("failed to marshal json: %v", err)
	}

	record := Record{
		Type:     godns.TypeToString[header.Rrtype],
		TTL:      header.Ttl,
		Name:     strings.TrimRight(header.Name, "."),
		Value:    data,
		stringer: answer.String,
	}

	// For MX records, extract priority.
	if mx, ok := answer.(*godns.MX); ok {
		record.Preference = &mx.Preference
	}

	return record, nil
}

func queryTypeFromNameserver(log *zap.Logger, nameserver, recordType, lookup string) ([]Record, error) {
	// Do the main DNS lookup.
	addr := nameserver + ":53"
	result, err := rawQuery(log, addr, godns.StringToType[recordType], lookup)
	if err != nil {
		log.Error("failed to lookup DNS record", zap.Error(err))
		return nil, err
	}

	records := []Record{}

	// Go through each answer and check if we need to do any traversals.
	answers := result.Answer
	if len(answers) == 0 && recordType == "NS" {
		answers = result.Ns
	}
	for _, answer := range answers {
		record, err := recordFromAnswer(answer)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

func findAuthoritativeNameserver(log *zap.Logger, hostname string) (string, RecordType, error) {
	// Select a root nameserver to begin our search
	rootNameserver := NextRootServer()

	resp := RecordType{}
	var recursiveSearch func(iteration int, nameserver string) (string, error)
	recursiveSearch = func(iteration int, nameserver string) (string, error) {
		if iteration > 10 {
			return "", errors.New("nameserver search depth exceeded")
		}
		iteration += 1

		msg, err := rawQuery(log, nameserver+":53", godns.TypeNS, hostname)
		if err != nil {
			return "", err
		}

		server := Server{
			Server:  nameserver,
			Records: []Record{},
		}

		// Add discovered answers/NSes to the records for showing to user
		for _, answer := range append(msg.Ns, msg.Answer...) {
			record, err := recordFromAnswer(answer)
			if err != nil {
				return "", err
			}

			server.Records = append(server.Records, record)
		}

		resp = append(resp, server)

		// Determine if we have further to traverse or if we've reached the end
		nsAnswers := []*godns.NS{}
		for _, answer := range msg.Answer {
			v, ok := answer.(*godns.NS)
			if ok {
				nsAnswers = append(nsAnswers, v)
			}
		}
		var cnameAnswer *godns.CNAME
		for _, answer := range msg.Answer {
			v, ok := answer.(*godns.CNAME)
			if ok {
				cnameAnswer = v
				break
			}
		}

		// If theres any answers of the NS type, we have found our authoritative
		// nameserver.
		if len(nsAnswers) > 0 {
			randomAnswer := nsAnswers[rand.Intn(len(nsAnswers))]
			return strings.TrimRight(randomAnswer.Ns, "."), nil
		}

		// If there's no NS type answers, but a cname answer, it means the user
		// has queried a cname. This is a weird behaviour.
		if cnameAnswer != nil {
			return "", nil
		}

		if len(msg.Ns) == 0 {
			// No answer, and no NS to follow. We've come to a dead end.
			return "", errors.New("no answer or authoritive server provided in dns response")
		}

		// We need to follow the nameservers deeper. Select a random one and
		// perform the search on that one now.
		switch v := msg.Ns[rand.Intn(len(msg.Ns))].(type) {
		case *godns.NS:
			return recursiveSearch(iteration, v.Ns)
		case *godns.SOA:
			return v.Ns, nil
		default:
			return "", fmt.Errorf("unexpected record returned: %T", v)
		}
	}

	authoritativeNameserver, err := recursiveSearch(0, rootNameserver)
	if err != nil {
		return "", nil, err
	}

	return authoritativeNameserver, resp, nil
}

func traceQuery(log *zap.Logger, dnsServer, recordType, hostname string) (Response, error) {
	authoritativeNameserver, answer, err := findAuthoritativeNameserver(log, hostname)
	if err != nil {
		return nil, err
	}

	// When tracing on a cname, we pick it up during the auth nameserver search
	// and aren't provided a authoritative nameserver to continue to.
	if authoritativeNameserver == "" {
		return Response{
			"TRACE": answer,
		}, nil
	}

	// Get the record types.
	recordTypes := []string{strings.ToUpper(recordType)}
	if recordType == "ANY" {
		// Not many DNS resolvers support this anymore, set it to literally all record types.
		recordTypes = []string{"A", "AAAA", "CNAME", "MX", "PTR", "SOA", "TXT"}
	} else if recordTypes[0] == "NS" {
		// We already have this data. Make this a blank slice.
		recordTypes = []string{}
	}

	eg := errgroup.Group{}
	answerLock := sync.Mutex{}
	// Spawn a goroutine to look up each record type.
	for _, recordLoop := range recordTypes {
		record := recordLoop
		eg.Go(func() error {
			records, err := queryTypeFromNameserver(log, authoritativeNameserver, record, hostname)
			if err != nil {
				return err
			}
			answerLock.Lock()
			answer[len(answer)-1].Records = append(
				answer[len(answer)-1].Records,
				records...,
			)

			answerLock.Unlock()
			return nil
		})
	}

	// Wait for the results.
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return Response{
		"TRACE": answer,
	}, nil
}

func recursiveQuery(log *zap.Logger, dnsServer, recordType, hostname string) (Response, error) {
	// Create the response map.
	responses := Response{}
	responsesLock := sync.Mutex{}

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
			records, err := queryTypeFromNameserver(log, dnsServer, record, hostname)
			if err != nil {
				return err
			}
			responsesLock.Lock()
			responses[record] = RecordType{
				Server{
					Server:  dnsServer,
					Records: records,
				},
			}
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

func Lookup(log *zap.Logger, dnsServer, recordType, hostname string, fullTrace bool) (Response, error) {
	// Add dot to hostname if necessary
	if !strings.HasSuffix(hostname, ".") {
		hostname += "."
	}

	if fullTrace {
		return traceQuery(log, dnsServer, recordType, hostname)
	}

	return recursiveQuery(log, dnsServer, recordType, hostname)
}

func reverseIP(ip net.IP) string {
	addressParts := strings.Split(ip.String(), ".")
	reversed := []string{}

	for i := len(addressParts) - 1; i >= 0; i-- {
		octet := addressParts[i]
		reversed = append(reversed, octet)
	}

	return strings.Join(reversed, ".")
}

func LookupRDNS(log *zap.Logger, ip net.IP, dnsServer string) (RecordType, error) {
	hostname := reverseIP(ip) + ".in-addr.arpa."
	resp, err := traceQuery(log, dnsServer, "PTR", hostname)
	if err != nil {
		return nil, err
	}

	return resp["TRACE"], nil
}
