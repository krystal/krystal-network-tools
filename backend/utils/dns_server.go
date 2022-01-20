package utils

import (
	"go.uber.org/zap"
	"os"
	"regexp"
)

// DNSServer is used to define the string that represents the DNS server. If the DNS_SERVER environment variable is set,
// it will be used as the DNS server. Otherwise, the default DNS server will be used.
var DNSServer string

var portRe = regexp.MustCompile(":[0-9]+$")

// InitializeDNSServer is used to initialize the DNS server.
func InitializeDNSServer(log *zap.Logger) {
	// Handle the environment variable override.
	DNSServer = os.Getenv("DNS_SERVER")
	if DNSServer != "" {
		// Check if a port is attached.
		if !portRe.MatchString(DNSServer) {
			DNSServer += ":53"
		}

		// Very poggers. Return here.
		return
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
	DNSServer = ns[len(ns)-1]
	if !portRe.MatchString(DNSServer) {
		DNSServer += ":53"
	}
}
