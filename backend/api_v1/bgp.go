package api_v1

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Defines the regex for a small and large bgp community.
var (
	smallCommunityRe = regexp.MustCompile(`\d+,\d+`)
	largeCommunityRe = regexp.MustCompile(`\d+, \d+, \d+`)
)

// bgpLine is a line of BGP data.
type bgpLine struct {
	// Code is used to define the BGP code.
	Code string

	// Line is used to store the BGP line.
	Line string

	// IsCont defines if this line is a continuation.
	IsCont bool
}

// Used to define the BGP type.
type bgpType int

const (
	_ = bgpType(iota)
	success
	table
	routeHeader
	routeType
	routeBgpAsPath
	routeBgpLocalPref
	routeBgpNextHop
	routeBgpCommunity
	routeBgpLargeCommunity
)

// Type is used to define the BGP type.
func (v *bgpLine) Type() (bgpType, string) {
	switch v.Code {
	case "0000":
		return success, v.Line
	case "1007":
		if strings.HasSuffix(v.Line, "Table ") {
			return table, v.Line
		}
		if strings.Contains(v.Line, "unreachable") {
			return routeHeader, strings.SplitN(v.Line, " ", 2)[0]
		}
	case "1008":
		if strings.HasPrefix(v.Line, "Type:") {
			return routeType, v.Line
		}
	case "1012":
		if strings.HasPrefix(v.Line, "BGP.as_path: ") {
			return routeBgpAsPath, v.Line[13:]
		}
		if strings.HasPrefix(v.Line, "BGP.local_pref: ") {
			return routeBgpLocalPref, v.Line[16:]
		}
		if strings.HasPrefix(v.Line, "BGP.next_hop: ") {
			return routeBgpNextHop, v.Line[14:]
		}
		if strings.HasPrefix(v.Line, "BGP.community: ") {
			return routeBgpCommunity, v.Line[15:]
		}
		if strings.HasPrefix(v.Line, "BGP.large_community: ") {
			return routeBgpLargeCommunity, v.Line[21:]
		}
	}

	// Dunno
	return 0, v.Line
}

// BGPRoute is used to define a route in BGP.
type BGPRoute struct {
	// Prefix is used to define the prefix of the BGP route.
	Prefix *string `json:"prefix"`

	// AsPath is set when the route has an AS path.
	AsPath []int `json:"as_path"`

	// LocalPref is set when the route has a local preference.
	LocalPref *int `json:"local_pref"`

	// NextHop is set when the route has a next hop.
	NextHop *string `json:"next_hop"`

	// Community is used to define a BGP community.
	Community []string `json:"community"`

	// LargeCommunity is used to define a large BGP community.
	LargeCommunity []string `json:"large_community"`
}

// BGPRouteSlice is used to define a slice of BGP routes that implements sort.Interface.
type BGPRouteSlice []*BGPRoute

// Len implements sort.Interface.
func (v BGPRouteSlice) Len() int {
	return len(v)
}

// Swap implements sort.Interface.
func (v BGPRouteSlice) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// Less implements sort.Interface.
func (v BGPRouteSlice) Less(i, j int) bool {
	x := v[i]
	y := v[j]
	if x.LocalPref != nil && y.LocalPref != nil && *x.LocalPref != *y.LocalPref {
		return *x.LocalPref > *y.LocalPref
	}
	return len(x.AsPath) < len(y.AsPath)
}

func makeBirdSocket() (io.ReadWriteCloser, error) {
	return net.Dial("unix", "/run/bird/bird.ctl")
}

func bgp(g group, socketBuilder func() (io.ReadWriteCloser, error)) {
	f := func(context *gin.Context) {
		// Get the IP address.
		ip, _ := url.PathUnescape(context.Param("ip"))
		rangeChunk := context.Param("range")
		if rangeChunk != "" {
			ip += "/" + rangeChunk
		}

		// Defines if this is JSON.
		isJson := context.ContentType() == "application/json"

		// Check the type of query we should make.
		queryType := ip
		if !ipRangeRegex.MatchString(ip) {
			if !ipRegex.MatchString(ip) {
				// In this situation, this isn't a valid IP address, and we should return.
				if isJson {
					context.JSON(400, map[string]string{
						"message": "Invalid IP address.",
					})
				} else {
					context.String(400, "Invalid IP address.")
				}
				return
			}
			queryType = "for " + queryType
		}

		// Make the socket.
		conn, err := socketBuilder()
		if err != nil {
			context.Error(err)
			return
		}
		defer conn.Close()

		// Allocate a 100KB page.
		b := make([]byte, 100*1024)

		// Drain the start sequence.
		_, err = conn.Read(b)
		if err != nil {
			context.Error(err)
			return
		}

		// Perform the query.
		query := "show route " + queryType + " all\n"
		_, err = conn.Write([]byte(query))
		if err != nil {
			context.Error(err)
			return
		}

		// Read the response.
		n, err := conn.Read(b)
		if err != nil {
			context.Error(err)
			return
		}
		b = append([]byte(nil), b[:n]...)

		// Now we are done with bird, close the connection.
		_ = conn.Close()

		// Check if this is a bird syntax error. This would mean that the IP address/range is not valid.
		// Note that no other errors are relevant here since we've already checked the IP address.
		if bytes.HasPrefix(b, []byte("9001 ")) {
			context.Error(&gin.Error{
				Err:  errors.New("bird lookup failed: " + string(b[5:])),
				Type: gin.ErrorTypePublic,
			})
			return
		}

		// If the content type isn't JSON, return it here.
		if !isJson {
			context.String(200, string(b))
			return
		}

		// Remove any new lines from hte slice.
		b = bytes.Trim(b, "\n\r")

		// Defines the lines.
		lines := bytes.Split(b, []byte("\n"))
		chunks := []bgpLine{}
		for _, v := range lines {
			if v[0] == ' ' {
				// Add a new line with the last code.
				lastCode := chunks[len(chunks)-1].Code
				chunks = append(chunks, bgpLine{
					Code:   lastCode,
					Line:   strings.Trim(string(v[1:]), " \t\r"),
					IsCont: true,
				})
				continue
			}

			// Get the code.
			code := string(v[:4])

			// Append to the chunk.
			chunks = append(chunks, bgpLine{
				Code: code,
				Line: strings.Trim(string(v[5:]), " \t\r"),
			})

			// If v[4] is a space, we should break.
			if v[4] == ' ' {
				break
			}
		}

		// Handle making all the routes.
		routes := BGPRouteSlice{}

		// Loop through the chunks to make each route.
		for _, v := range chunks {
			switch type_, clean := v.Type(); type_ {
			case routeHeader:
				prefix := clean
				if prefix != "" {
					if len(routes) != 0 {
						p := routes[len(routes)-1].Prefix
						if p != nil {
							prefix = *p
						}
					}
				}
				routes = append(routes, &BGPRoute{Prefix: &prefix})
			case routeBgpAsPath:
				x := make([]int, 0)
				for _, v := range strings.Split(clean, " ") {
					i, err := strconv.Atoi(v)
					if err == nil {
						x = append(x, i)
					}
				}
				routes[len(routes)-1].AsPath = x
			case routeBgpLocalPref:
				x, err := strconv.Atoi(clean)
				if err != nil {
					context.Error(err)
					return
				}
				routes[len(routes)-1].LocalPref = &x
			case routeBgpNextHop:
				routes[len(routes)-1].NextHop = &clean
			case routeBgpCommunity:
				a := smallCommunityRe.FindAllString(clean, -1)
				if a == nil {
					context.Error(errors.New("string not correct format for bgp community"))
					return
				}
				routes[len(routes)-1].Community = a
			case routeBgpLargeCommunity:
				a := largeCommunityRe.FindAllString(clean, -1)
				if a == nil {
					context.Error(errors.New("string not correct format for bgp large community"))
					return
				}
				routes[len(routes)-1].LargeCommunity = a
			}
		}

		// Sort the slice.
		sort.Sort(routes)

		// Return the routes.
		context.JSON(200, routes)
	}
	g.GET("/:ip", f)
	g.GET("/:ip/:range", f)
}
