package dns

import (
	"go.uber.org/zap"
	"os"
)

// GetCachedDNSServer is used to get the cache DNS server.
func GetCachedDNSServer(log *zap.Logger) string {
	// Handle the environment variable override.
	s := os.Getenv("DNS_SERVER")
	if s != "" {
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
	return s
}
