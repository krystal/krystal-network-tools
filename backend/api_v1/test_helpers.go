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

	handlers map[string][]gin.HandlerFunc
	method   map[string]string
}

func (m *mockHandler) set(method string, path string, handlers []gin.HandlerFunc) gin.IRoutes {
	if m.method == nil {
		m.method = map[string]string{}
	}
	m.method[path] = method
	if m.handlers == nil {
		m.handlers = map[string][]gin.HandlerFunc{}
	}
	m.handlers[path] = handlers
	return m
}

func (m *mockHandler) GET(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return m.set("GET", relativePath, handlers)
}

func mockGroupMultiHn(t *testing.T, paths []string, methods map[string]string, f func(group)) map[string]gin.HandlerFunc {
	t.Helper()
	m := mockHandler{}
	f(&m)
	if len(m.handlers) != len(paths) {
		t.Error(len(m.handlers), "routes set")
		return nil
	} else {
		newMap := map[string]gin.HandlerFunc{}
		for _, path := range paths {
			method := methods[path]
			resultMethod, ok := m.method[path]
			if ok {
				assert.Equal(t, method, resultMethod)
				hn := m.handlers[path]
				switch len(hn) {
				case 0:
					t.Error("no handlers set")
					return nil
				case 1:
					newMap[path] = hn[0]
				default:
					t.Error("more than 1 handler is not supported")
					return nil
				}
			} else {
				t.Error("no handler set for", path)
			}
		}
		return newMap
	}
}

func mockGroupSingleHn(t *testing.T, method, path string, f func(group)) gin.HandlerFunc {
	t.Helper()
	m := mockGroupMultiHn(t, []string{path}, map[string]string{path: method}, f)
	if m == nil {
		return nil
	}
	return m[path]
}
