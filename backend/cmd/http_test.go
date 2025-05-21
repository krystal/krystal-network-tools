package cmd

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	r = gin.New()
	r.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusBadRequest)
		c.Error(&gin.Error{
			Type: gin.ErrorTypePublic,
			Err:  errors.New("test error"),
		})
	})

}

func TestErrorHandlingMiddleware(t *testing.T) {
	w := httptest.NewRecorder()

	r.Use(errorHandling)
	req, _ := http.NewRequest(
		"GET", "/ping",
		nil,
	)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCorsMiddleware(t *testing.T) {
	w := httptest.NewRecorder()

	r.Use(cors)
	req, _ := http.NewRequest(
		"OPTIONS", "/ping",
		nil,
	)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Content-Type", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "GET, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
}

func TestJsonMiddleware(t *testing.T) {
	w := httptest.NewRecorder()

	r.Use(json)
	req, _ := http.NewRequest(
		"POST", "/ping",
		nil,
	)
	req.Header.Set("Content-Type", "application/json+vnd.v1")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "{\"error\":\"Content-Type must be application/json\"}", w.Body.String())
}
