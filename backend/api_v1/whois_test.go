package api_v1

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jimeh/go-golden"
	"github.com/stretchr/testify/assert"
)

const (
	resultHeader = "RESULT HEADER - DO NOT MANUALLY EDIT THIS LINE"
	errorHeader  = "ERROR HEADER - DO NOT MANUALLY EDIT THIS LINE"
)

func makeWhoisResultFile(t *testing.T, result string, err error) {
	if err == nil {
		golden.SetP(t, "whois_result", []byte(resultHeader+"\n"+result))
	} else {
		golden.SetP(t, "whois_result", []byte(errorHeader+"\n"+err.Error()))
	}
}

func readWhoisResultFile(t *testing.T) (string, error) {
	b := golden.GetP(t, "whois_result")
	a := bytes.SplitN(b, []byte("\n"), 2)
	if len(a) == 1 {
		return "", errors.New("header invalid")
	}
	switch string(a[0]) {
	case resultHeader:
		return string(a[1]), nil
	case errorHeader:
		return "", errors.New(string(a[1]))
	default:
		return "", errors.New("header invalid")
	}
}

type goldenWhoisWriter struct {
	t        *testing.T
	lookuper whoisLookuper
}

func (g goldenWhoisWriter) Whois(hostOrIp string) (string, error) {
	r, err := g.lookuper.Whois(hostOrIp)
	makeWhoisResultFile(g.t, r, err)
	return r, err
}

type goldenWhoisLookuper struct {
	t *testing.T
}

func (g goldenWhoisLookuper) Whois(hostOrIp string) (string, error) {
	return readWhoisResultFile(g.t)
}

func Test_whois(t *testing.T) {
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
		{
			name: "lookup fail text",
			code: http.StatusBadRequest,
			json: false,
			addr: "",
		},
		{
			name: "lookup fail json",
			code: http.StatusBadRequest,
			json: true,
			addr: "",
		},
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
			hn := mockGroupSingleHn(t, "GET", "/:hostOrIp", func(g group) {
				if golden.Update() {
					whois(g, goldenWhoisWriter{
						t:        t,
						lookuper: defaultWhoisLookuper{},
					})
				} else {
					whois(g, goldenWhoisLookuper{t})
				}
			})
			if hn == nil {
				return
			}
			hn(c)
			assert.Equal(t, tt.code, w.Code)
			body := w.Body.String()
			if golden.Update() {
				golden.SetP(t, "result", []byte(body))
			}
			assert.Equal(t, string(golden.GetP(t, "result")), body)
		})
	}
}
