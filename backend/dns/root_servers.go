package dns

import "sync/atomic"

// RootServers is a slice of root servers.
var RootServers []string

// Create the slice.
func init() {
	chars := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m"}
	RootServers = make([]string, len(chars))
	for i, c := range chars {
		RootServers[i] = c + ".root-servers.net."
	}
}

// Defines the root server index.
var rootServerIndex uintptr

// NextRootServer returns the next root server along, looping back around.
func NextRootServer() string {
	new_ := atomic.AddUintptr(&rootServerIndex, 1)
	if new_ >= uintptr(len(RootServers)) {
		// In the event we hit this, we should wrap around to zero.
		// We may hit zero a couple times with this, but it's close enough.
		atomic.StoreUintptr(&rootServerIndex, 0)
		return RootServers[0]
	}
	return RootServers[new_]
}
