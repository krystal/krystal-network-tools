package api_v1

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Test_userIp(t *testing.T) {
	// Allow the function to insert into the group as it normally would.
	hn := mockGroupSingleHn(t, "GET", "/ip", userIp)
	if hn == nil {
		return
	}

	// Run the tests on the handler.
	tests := []struct {
		name string

		json    bool
		expects string
	}{
		{
			name:    "text",
			json:    false,
			expects: "1.1.1.1",
		},
		{
			name:    "json",
			json:    true,
			expects: `{"ip":"1.1.1.1"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = &http.Request{
				RemoteAddr: "1.1.1.1:443",
				URL:        &url.URL{Path: "/ip"},
				Header:     http.Header{},
			}
			if tt.json {
				c.Request.Header.Set("Content-Type", "application/json")
			}
			hn(c)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.expects, w.Body.String())
		})
	}
}
