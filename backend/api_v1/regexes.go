package api_v1

import "regexp"

//nolint:lll
const ipRangePrefix = `^(((([1]?\d)?\d|2[0-4]\d|25[0-5])\.){3}(([1]?\d)?\d|2[0-4]\d|25[0-5]))|([\da-fA-F]{1,4}(\:[\da-fA-F]{1,4}){7})|(([\da-fA-F]{1,4}:){0,5}::([\da-fA-F]{1,4}:){0,5}[\da-fA-F]{1,4})`

var (
	// Defines regex for getting a IP range.
	ipRangeRegex = regexp.MustCompile(ipRangePrefix + `(\/([0-9]|[1-2][0-9]|3[0-2]))?$`)

	// Defines regex for getting a IP.
	ipRegex = regexp.MustCompile(ipRangePrefix + "$")
)
