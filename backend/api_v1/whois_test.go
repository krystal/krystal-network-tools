package api_v1

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jimeh/go-golden"
	"github.com/stretchr/testify/assert"
)

var (
	queryTimeRegex  = regexp.MustCompile(`Query time: \d+ msec`)
	whenRegex       = regexp.MustCompile(`WHEN: .+ GMT`)
	lastUpdateRegex = regexp.MustCompile(`Last update of WHOIS database: .+Z`)
	queryInfoRegex  = regexp.MustCompile("% This query was served by the RIPE " +
		"Database Query Service version 1.102.2 (.+)")
)

func Test_whois(t *testing.T) {
	// Allow the function to insert into the group as it normally would.
	hn := mockGroup(t, "GET", "/:hostOrIp", whois)
	if hn == nil {
		return
	}

	// Run the tests on the handler.
	tests := []struct {
		name string

		code int
		json bool
		addr string
	}{
		{
			name: "ip success text",
			code: http.StatusOK,
			json: false,
			addr: "81.2.115.158",
		},
		{
			name: "ip success json",
			code: http.StatusOK,
			json: true,
			addr: "81.2.115.158",
		},
		{
			name: "hostname success text",
			code: http.StatusOK,
			json: false,
			addr: "one.one.one.one",
		},
		{
			name: "hostname success json",
			code: http.StatusOK,
			json: true,
			addr: "one.one.one.one",
		},
		// TODO: tests for failures
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = &http.Request{
				URL:    &url.URL{Path: "/" + tt.addr},
				Header: http.Header{},
			}
			c.Params = append(c.Params, gin.Param{Key: "hostOrIp", Value: tt.addr})
			if tt.json {
				c.Request.Header.Set("Content-Type", "application/json")
			}
			hn(c)
			assert.Equal(t, tt.code, w.Code)
			v := queryTimeRegex.ReplaceAllString(w.Body.String(), "")
			v = queryInfoRegex.ReplaceAllString(v, "")
			v = lastUpdateRegex.ReplaceAllString(v, "")
			v = whenRegex.ReplaceAllString(v, "")
			if golden.Update() {
				golden.Set(t, []byte(v))
			}
			assert.Equal(t, string(golden.Get(t)), v)
		})
	}
}
