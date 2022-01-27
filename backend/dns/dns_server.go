package dns

import (
	"go.uber.org/zap"
	"os"
	"regexp"
)

var portRe = regexp.MustCompile(":[0-9]+$")

// GetDNSServer is used to get the DNS server.
func GetDNSServer(log *zap.Logger) string {
	// Handle the environment variable override.
	s := os.Getenv("DNS_SERVER")
	if s != "" {
		// Check if a port is attached.
		if !portRe.MatchString(s) {
			s += ":53"
		}

		// Very poggers. Return here.
		return s
	}

	// Get the systems default DNS server.
	resolv, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		panic(err)
	}

	// Find the nameserver.
	log.Warn("No DNS_SERVER environment variable set. Using system default DNS server. In production, " +
		"this should be set.")
	ns := getNameservers(resolv)
	if len(ns) == 0 {
		panic("no DNS server found")
	}
	s = ns[len(ns)-1]
	if !portRe.MatchString(s) {
		s += ":53"
	}
	return s
}
