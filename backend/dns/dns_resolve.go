package dns

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
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
	return "TODO"
}

type RecordType []Server

func (rt RecordType) String() string {
	return "TODO"
}

type Server struct {
	Server  string   `json:"server"`
	Records []Record `json:"records"`
}

func (srv Server) String() string {
	str := srv.Server + ": \n"
	for _, record := range srv.Records {
		if record.stringer != nil {
			str += record.stringer()
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
		header := answer.Header()

		// Get the data from the record.
		// Due to the nature of the library, this is sadly a little magical.
		var data []byte
		reflectValue := reflect.Indirect(reflect.ValueOf(answer))
		reflectType := reflectValue.Type()
		n := reflectType.NumField()
		if cname, ok := answer.(*godns.CNAME); ok {
			data, _ = json.Marshal(cname.Target)
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
				Value:      answer,
				RemoveKeys: []string{"Hdr"},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to marshal json: %v", err)
			}
		}

		record := Record{
			Type:     godns.TypeToString[header.Rrtype],
			TTL:      header.Ttl,
			Name:     strings.TrimRight(lookup, "."),
			Value:    data,
			stringer: answer.String,
		}

		// For MX records, extract priority.
		if mx, ok := answer.(*godns.MX); ok {
			record.Preference = &mx.Preference
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
		msg, err := rawQuery(log, nameserver+":53", godns.TypeNS, hostname)
		if err != nil {
			return "", err
		}

		server := Server{
			Server:  nameserver,
			Records: []Record{},
		}

		// Add discovered answers/NSes to the records for showing to user
		for _, record := range append(msg.Answer, msg.Ns...) {
			b, _ := json.Marshal(record)
			header := record.Header()
			server.Records = append(server.Records, Record{
				Type:     godns.TypeToString[header.Rrtype],
				Name:     strings.TrimRight(header.Name, "."),
				Value:    b,
				TTL:      record.Header().Ttl,
				stringer: record.String,
			})
		}

		resp = append(resp, server)

		// Determine if we have further to traverse or if we've reached the end

		if len(msg.Answer) > 0 {
			authoritativeNameserver := msg.Answer[rand.Intn(len(msg.Answer))]

			switch t := authoritativeNameserver.(type) {
			case *godns.NS:
				return strings.TrimRight(t.Ns, "."), nil
			case *godns.CNAME:
				return nameserver, nil
			default:
				return "", fmt.Errorf(
					"unexpected godns type: %T", authoritativeNameserver,
				)
			}
		}

		if len(msg.Ns) == 0 {
			// If it doesn't return a answer, or somewhere we can go for an answer
			// we are effectively lost !
			return "", errors.New("no answer or authoritive server provided in dns response")
		}

		nextNameserver := msg.Ns[rand.Intn(len(msg.Ns))]

		nameserverRecord, ok := nextNameserver.(*godns.NS)
		if !ok {
			return "", fmt.Errorf("unexpected record returned: %T", nextNameserver)
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

func traceQuery(log *zap.Logger, dnsServer, recordType, hostname string) (Response, error) {
	// Create the response map.
	responses := Response{}

	authoritativeNameserver, answer, err := findAuthoritativeNameserver(log, hostname)
	if err != nil {
		return nil, err
	}

	answerLock := sync.Mutex{}

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

	responses["TRACE"] = answer

	return responses, nil
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
