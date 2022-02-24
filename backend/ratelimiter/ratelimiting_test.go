package ratelimiter

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
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
	prefix *regexp.Regexp
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

func (r request) do(t *testing.T, f func(bucketContext), wg *sync.WaitGroup) {
	innerDo := func() {
		defer func() {
			if r.parallel {
				wg.Done()
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
							if formatter.prefix == nil {
								assert.Equal(t, formatter.value, v)
							} else {
								formatted := fmt.Sprint(v)
								if !formatter.prefix.MatchString(formatted) {
									t.Errorf("value %s does not match regex prefix", formatted)
								}
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
		// Tests that the buckets can handle a significant number of requests
		// one after another. Each request should not be ratelimited.
		{
			name:    "non-parallel requests",
			maxUses: 1000,
			per:     time.Second,
			backoff: time.Second,
			reqs:    makeGenericRequests(1000, false),
		},

		// Tests that the buckets can handle many requests coming into the
		// web server at once. Each request should not be ratelimited.
		{
			name:    "parallel requests",
			maxUses: 1000,
			per:     time.Second,
			backoff: time.Second,
			reqs:    makeGenericRequests(1000, true),
		},

		// Check that ratelimits are hit and returned properly when not in
		// parallel. This allows us to isolate issues with parallelism and
		// matching ratelimits.
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
								prefix: regexp.MustCompile("^99[0-9] milliseconds "),
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

		// Tests that when requests are run after we are ratelimited,
		// we return the ratelimit properly. This allows us to differentiate
		// issues with race from the test above.
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
								prefix: regexp.MustCompile("^99[0-9] milliseconds "),
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

		// Test that after the backoff time, requests are successful again.
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
								prefix: regexp.MustCompile("^[1-4] milliseconds "),
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

		// Test multiple buckets with different backoffs and request speeds.
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
								prefix: regexp.MustCompile("^[1-4] milliseconds "),
							},
						},
					},
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
			b := newBucket(zaptest.NewLogger(t), tt.maxUses, tt.per, tt.backoff)
			wg := &sync.WaitGroup{}
			parallelTasksBefore := false
			for _, v := range tt.reqs {
				if v.parallel {
					// If these tests are parallel, we want to add to the wait group.
					// We also set parallelTasksBefore to true. This is done so that if we have a
					// non-parallel task next, we know from it that there are parallel tasks
					// running and to wait.
					wg.Add(1)
					parallelTasksBefore = true
				} else if parallelTasksBefore {
					// If there were parallel tasks before this one (and we established this one is
					// not parallel since this is an else if), we should wait for the parallel
					// tasks to complete and then say this is not parallel. This prevents a situation
					// where non-parallel tasks run at the same time as parallel ones.
					parallelTasksBefore = false
					wg.Wait()
				}

				// Launch the request.
				v.do(t, b, wg)
			}

			// Wait here in case any parallel tasks are still running.
			wg.Wait()
		})
	}
}
