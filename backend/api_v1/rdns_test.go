package api_v1

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Test_rdns(t *testing.T) {
	// Allow the function to insert into the group as it normally would.
	hn := mockGroupSingleHn(t, "GET", "/:ip", rdns)
	if hn == nil {
		return
	}

	// Run the tests on the handler.
	tests := []struct {
		name string

		code    int
		json    bool
		addr    string
		expects string
	}{
		{
			name:    "success text",
			code:    http.StatusOK,
			json:    false,
			addr:    "1.1.1.1",
			expects: "one.one.one.one.",
		},
		{
			name:    "success json",
			code:    http.StatusOK,
			json:    true,
			addr:    "1.1.1.1",
			expects: `{"hostname":"one.one.one.one."}`,
		},
		{
			name:    "error text",
			code:    http.StatusBadRequest,
			json:    false,
			addr:    "bad_ip",
			expects: "Failed to find IP",
		},
		{
			name:    "error json",
			code:    http.StatusBadRequest,
			json:    true,
			addr:    "bad_ip",
			expects: `{"message":"Failed to find IP"}`,
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
			c.Params = append(c.Params, gin.Param{Key: "ip", Value: tt.addr})
			if tt.json {
				c.Request.Header.Set("Content-Type", "application/json")
			}
			hn(c)
			assert.Equal(t, tt.code, w.Code)
			assert.Equal(t, tt.expects, w.Body.String())
		})
	}
}
