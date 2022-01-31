package api_v1

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"testing"
)

type group interface {
	GET(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes
}

type mockHandler struct {
	// NOT here for actual usage. Solely here to satisfy the interface.
	gin.IRoutes

	handlers []gin.HandlerFunc
	method   string
	path     string
}

func (m *mockHandler) set(method string, path string, handlers []gin.HandlerFunc) gin.IRoutes {
	m.method = method
	m.path = path
	m.handlers = handlers
	return m
}

func (m *mockHandler) GET(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return m.set("GET", relativePath, handlers)
}

func mockGroup(t *testing.T, method, path string, f func(group)) gin.HandlerFunc {
	t.Helper()
	m := mockHandler{}
	f(&m)
	if m.method == "" {
		t.Error("no route set")
	} else {
		assert.Equal(t, method, m.method)
		assert.Equal(t, path, m.path)
		switch len(m.handlers) {
		case 0:
			t.Error("no handlers set")
			return nil
		case 1:
			return m.handlers[0]
		default:
			t.Error("more than 1 handler is not supported")
			return nil
		}
	}
	return nil
}
