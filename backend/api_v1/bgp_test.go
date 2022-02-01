package api_v1

import (
	"encoding/json"
	"errors"
	"github.com/jimeh/go-golden"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func attemptJsonBeautify(b []byte) []byte {
	var x interface{}
	err := json.Unmarshal(b, &x)
	if err != nil {
		return b
	}
	y, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		return b
	}
	return y
}

type bgpTape struct {
	// Defines the test object.
	t *testing.T

	// Defines the decoded tape for when it is loaded.
	decodedTape []string

	// Defines the index on reads.
	readIndex int

	// Defines the writes.
	writes []string

	// Defines the index on writes.
	writeIndex int

	// Defines if the tape is closed.
	closed bool
}

func newBgpTape(t *testing.T, tapeFilename string, writes []string) *bgpTape {
	// Find the file path to the file.
	_, filename, _, _ := runtime.Caller(1)
	filename = filepath.Join(filepath.Dir(filename), "bgp_tapes", tapeFilename)

	// Load the tape.
	b, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var decodedTape []string
	if err = json.Unmarshal(b, &decodedTape); err != nil {
		t.Fatal(err)
	}

	// Return the tape.
	return &bgpTape{
		t:           t,
		decodedTape: decodedTape,
		writes:      writes,
	}
}

func (t *bgpTape) Read(p []byte) (n int, err error) {
	t.t.Helper()
	if t.readIndex == len(t.decodedTape) {
		t.t.Error("read too far")
		return 0, io.EOF
	}
	n = copy(p, t.decodedTape[t.readIndex])
	t.readIndex++
	return
}

func (t *bgpTape) Write(data []byte) (n int, err error) {
	t.t.Helper()
	if t.writeIndex == len(t.writes) {
		t.t.Error("write too far")
		return 0, io.EOF
	}
	assert.Equal(t.t, t.writes[t.writeIndex], string(data))
	return
}

func (t *bgpTape) Close() error {
	t.closed = true
	return nil
}

func Test_bgp(t *testing.T) {
	tests := []struct {
		name string

		tapeFile    string
		writes      []string
		socketError string
		code        int
		addr        string
		json        bool
		urlEncode   bool
	}{
		{
			name:        "socket error",
			tapeFile:    "1111.json",
			addr:        "1.1.1.1",
			socketError: "bad",
		},
		{
			name:     "successful ip json",
			tapeFile: "1111.json",
			writes: []string{
				"show route for 1.1.1.1 all\n",
			},
			code: http.StatusOK,
			addr: "1.1.1.1",
			json: true,
		},
		{
			name:     "successful ip text",
			tapeFile: "1111.json",
			writes: []string{
				"show route for 1.1.1.1 all\n",
			},
			code: http.StatusOK,
			addr: "1.1.1.1",
			json: false,
		},
		{
			name:     "successful un-encoded range json",
			tapeFile: "8880_24.json",
			writes: []string{
				"show route 8.8.8.0/24 all\n",
			},
			code: http.StatusOK,
			addr: "8.8.8.0/24",
			json: true,
		},
		{
			name:     "successful un-encoded range text",
			tapeFile: "8880_24.json",
			writes: []string{
				"show route 8.8.8.0/24 all\n",
			},
			code: http.StatusOK,
			addr: "8.8.8.0/24",
			json: true,
		},
		{
			name:     "successful encoded range json",
			tapeFile: "8880_24.json",
			writes: []string{
				"show route 8.8.8.0/24 all\n",
			},
			code:      http.StatusOK,
			addr:      "8.8.8.0/24",
			urlEncode: true,
			json:      true,
		},
		{
			name:     "successful encoded range text",
			tapeFile: "8880_24.json",
			writes: []string{
				"show route 8.8.8.0/24 all\n",
			},
			code:      http.StatusOK,
			addr:      "8.8.8.0/24",
			urlEncode: true,
			json:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the mocker.
			mocker := newBgpTape(t, tt.tapeFile, tt.writes)

			// Allow the function to insert into the group as it normally would.
			handlers := mockGroupMultiHn(t, []string{"/:ip", "/:ip/:range"}, map[string]string{
				"/:ip": "GET", "/:ip/:range": "GET",
			}, func(g group) {
				bgp(g, func() (io.ReadWriteCloser, error) {
					if tt.socketError != "" {
						return nil, errors.New(tt.socketError)
					}
					return mocker, nil
				})
			})
			if handlers == nil {
				return
			}

			// Call the correct handler.
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			var s []string
			if !tt.urlEncode {
				s = strings.SplitN(tt.addr, "/", 2)
			}
			var path string
			hnPath := "/:ip"
			if tt.urlEncode {
				v := url.PathEscape(tt.addr)
				path = "/" + v
				c.Params = append(c.Params, gin.Param{Key: "ip", Value: v})
			} else {
				c.Params = append(c.Params, gin.Param{Key: "ip", Value: s[0]})
				path = "/" + s[0]
				if len(s) != 1 {
					hnPath += "/:range"
					path += "/" + s[1]
					c.Params = append(c.Params, gin.Param{Key: "range", Value: s[1]})
				}
			}
			c.Request = &http.Request{
				URL:    &url.URL{Path: path},
				Header: http.Header{},
			}
			if tt.json {
				c.Request.Header.Set("Content-Type", "application/json")
			}
			handlers[hnPath](c)

			// Handle socket errors.
			if tt.socketError == "" {
				assert.Equal(t, tt.code, w.Code)
				if golden.Update() {
					golden.Set(t, attemptJsonBeautify(w.Body.Bytes()))
				}
				assert.Equal(t, string(golden.Get(t)), string(attemptJsonBeautify(w.Body.Bytes())))
			} else {
				var err error
				if len(c.Errors) != 0 {
					err = c.Errors[0]
				}
				assert.EqualError(t, err, tt.socketError)
			}
		})
	}
}
