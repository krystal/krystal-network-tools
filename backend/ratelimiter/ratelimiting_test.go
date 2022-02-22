package ratelimiter

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

type jsonValue struct {
	status int
	keys   []string
}

type argValue struct {
	prefix string
	value  interface{}
}

type stringfValue struct {
	status int
	format string
	values []interface{}
}

type stringfCmpValue struct {
	status int
	format string
	values []argValue
}

type jsonOrStringfValue = interface{}

type mocker struct {
	// Defines the testing object.
	t *testing.T

	// Defines the result.
	result jsonOrStringfValue

	// Defines the client IP.
	clientIP string

	// Defines the content type.
	contentType string

	// Defines if this was aborted.
	aborted bool
}

func (m *mocker) JSON(status int, obj interface{}) {
	v := reflect.ValueOf(obj)
	keysReflect := v.MapKeys()
	keys := make([]string, len(keysReflect))
	for i, key := range keysReflect {
		keys[i] = key.String()
	}
	m.result = jsonValue{status, keys}
}

func (m *mocker) String(status int, format string, values ...interface{}) {
	m.result = stringfValue{status, format, values}
}

func (m *mocker) FullPath() string {
	return "/demo/:part"
}

func (m *mocker) ClientIP() string {
	return m.clientIP
}

func (m *mocker) ContentType() string {
	return m.contentType
}

func (m *mocker) Abort() {
	m.aborted = true
}

func (m *mocker) Value(key interface{}) interface{} {
	m.t.Helper()
	if key != 0 {
		m.t.Error("invalid Value key")
	}
	return &http.Request{URL: &url.URL{Path: "/demo/path"}}
}

var _ bucketContext = (*mocker)(nil)

type request struct {
	jsonValue    *jsonValue
	stringfValue *stringfCmpValue
	sleep        time.Duration
	contentType  string
	clientIP     string
	parallel     bool
}

func (r request) do(t *testing.T, f func(bucketContext), cb func()) {
	innerDo := func() {
		defer func() {
			if cb != nil {
				cb()
			}
		}()
		t.Helper()
		m := &mocker{
			t:           t,
			clientIP:    r.clientIP,
			contentType: r.contentType,
		}
		f(m)
		if r.jsonValue == nil && r.stringfValue == nil {
			// In this case, it shouldn't be aborted.
			if m.aborted {
				t.Error("aborted for nil value")
			}
			if m.result != nil {
				t.Error("unexpected result sent")
			}
		} else {
			if !m.aborted {
				t.Error("not aborted for non-nil response")
			}
			if m.result == nil {
				t.Error("nil result sent")
			} else {
				if r.stringfValue == nil {
					y := m.result.(jsonValue)
					assert.Equal(t, r.jsonValue.status, y.status)
					assert.ElementsMatch(t, r.jsonValue.keys, y.keys)
				} else {
					y := m.result.(stringfValue)
					x := r.stringfValue
					assert.Equal(t, x.status, y.status)
					assert.Equal(t, x.format, y.format)
					if len(x.values) == len(y.values) {
						for i, v := range y.values {
							formatter := x.values[i]
							if formatter.prefix == "" {
								assert.Equal(t, formatter.value, v)
							} else {
								assert.True(t, strings.HasPrefix(fmt.Sprint(v), formatter.prefix))
							}
						}
					} else {
						t.Error("invalid number of values")
					}
				}
			}
		}
	}
	if r.parallel {
		time.AfterFunc(r.sleep, innerDo)
	} else {
		if r.sleep != 0 {
			time.Sleep(r.sleep)
		}
		innerDo()
	}
}

func makeGenericRequests(n int, parallel bool) []request {
	a := make([]request, n)
	for i := range a {
		a[i] = request{
			contentType: "text/plain",
			clientIP:    "1.1.1.1",
			parallel:    parallel,
		}
	}
	return a
}

func TestNewBucket(t *testing.T) {
	tests := []struct {
		name string

		reqs    []request
		maxUses uint64
		per     time.Duration
		backoff time.Duration
	}{
		{
			name:    "non-parallel requests",
			maxUses: 1000,
			per:     time.Second,
			backoff: time.Second,
			reqs:    makeGenericRequests(1000, false),
		},
		{
			name:    "parallel requests",
			maxUses: 1000,
			per:     time.Second,
			backoff: time.Second,
			reqs:    makeGenericRequests(1000, true),
		},
		{
			name:    "non-parallel ratelimit hits",
			maxUses: 2,
			per:     time.Second,
			backoff: time.Second,
			reqs: []request{
				{
					contentType: "text/plain",
					clientIP:    "1.1.1.1",
					parallel:    false,
				},
				{
					contentType: "application/json",
					clientIP:    "1.1.1.1",
					parallel:    false,
				},
				{
					contentType: "text/plain",
					clientIP:    "1.1.1.1",
					parallel:    false,
					stringfValue: &stringfCmpValue{
						status: 429,
						format: "You have been ratelimited! Try again in %s.",
						values: []argValue{
							{
								prefix: "999 milliseconds ",
							},
						},
					},
				},
				{
					contentType: "application/json",
					clientIP:    "1.1.1.1",
					parallel:    false,
					jsonValue: &jsonValue{
						status: 429,
						keys:   []string{"wait_ms", "message"},
					},
				},
			},
		},
		{
			name:    "parallel ratelimit hits",
			maxUses: 2,
			per:     time.Second,
			backoff: time.Second,
			reqs: []request{
				{
					contentType: "text/plain",
					clientIP:    "1.1.1.1",
					parallel:    false,
				},
				{
					contentType: "application/json",
					clientIP:    "1.1.1.1",
					parallel:    false,
				},
				{
					contentType: "text/plain",
					clientIP:    "1.1.1.1",
					parallel:    true,
					stringfValue: &stringfCmpValue{
						status: 429,
						format: "You have been ratelimited! Try again in %s.",
						values: []argValue{
							{
								prefix: "999 milliseconds ",
							},
						},
					},
				},
				{
					contentType: "application/json",
					clientIP:    "1.1.1.1",
					parallel:    true,
					jsonValue: &jsonValue{
						status: 429,
						keys:   []string{"wait_ms", "message"},
					},
				},
			},
		},
		{
			name:    "backoff",
			maxUses: 2,
			per:     time.Millisecond,
			backoff: time.Millisecond * 5,
			reqs: []request{
				{
					contentType: "text/plain",
					clientIP:    "1.1.1.1",
					parallel:    false,
				},
				{
					contentType: "application/json",
					clientIP:    "1.1.1.1",
					parallel:    false,
				},
				{
					contentType: "text/plain",
					clientIP:    "1.1.1.1",
					parallel:    false,
					stringfValue: &stringfCmpValue{
						status: 429,
						format: "You have been ratelimited! Try again in %s.",
						values: []argValue{
							{
								prefix: "4 milliseconds ",
							},
						},
					},
				},
				{
					contentType: "application/json",
					clientIP:    "1.1.1.1",
					sleep:       time.Millisecond * 6,
					parallel:    false,
				},
			},
		},
		{
			name:    "multiple buckets",
			maxUses: 2,
			per:     time.Millisecond,
			backoff: time.Millisecond * 5,
			reqs: []request{
				{
					contentType: "text/plain",
					clientIP:    "1.1.1.1",
					parallel:    false,
				},
				{
					contentType: "application/json",
					clientIP:    "8.8.8.8",
					parallel:    false,
				},
				{
					contentType: "application/json",
					clientIP:    "1.1.1.1",
					parallel:    false,
				},
				{
					contentType: "text/plain",
					clientIP:    "1.1.1.1",
					parallel:    false,
					stringfValue: &stringfCmpValue{
						status: 429,
						format: "You have been ratelimited! Try again in %s.",
						values: []argValue{
							{
								prefix: "4 milliseconds ",
							},
						},
					},
				},
				{
					contentType: "application/json",
					clientIP:    "8.8.8.8",
					parallel:    false,
				},
				{
					contentType: "application/json",
					clientIP:    "8.8.8.8",
					sleep:       time.Millisecond,
					parallel:    true,
				},
				{
					contentType: "application/json",
					clientIP:    "1.1.1.1",
					sleep:       time.Millisecond * 6,
					parallel:    true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, bt := newBucket(zaptest.NewLogger(t), tt.maxUses, tt.per, tt.backoff, true)
			wg := sync.WaitGroup{}
			parallelTasks := false
			for _, v := range tt.reqs {
				var f func()
				if v.parallel {
					wg.Add(1)
					f = wg.Done
					parallelTasks = true
				} else if parallelTasks {
					// Wait for the tasks to be done first.
					parallelTasks = false
					wg.Wait()
				}
				v.do(t, b, f)
			}
			wg.Wait()
			for _, v := range bt.timers {
				v.Stop()
			}
		})
	}
}
