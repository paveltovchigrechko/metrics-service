package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/paveltovchigrechko/metrics-service/internal/model"
	"github.com/paveltovchigrechko/metrics-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendEscaped(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"Alloc", "Alloc"},
		{"<script>", "&lt;script&gt;"},
		{`a&b`, "a&amp;b"},
		{`"quoted"`, "&#34;quoted&#34;"},
		{"it's", "it&#39;s"},
		{"<>&\"'", "&lt;&gt;&amp;&#34;&#39;"},
		{"ключ<1>", "ключ&lt;1&gt;"},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, string(appendEscaped(nil, tc.in)))
		})
	}

	t.Run("appends to existing data", func(t *testing.T) {
		assert.Equal(t, "x:&lt;", string(appendEscaped([]byte("x:"), "<")))
	})
}

func TestAppendValue(t *testing.T) {
	tests := []struct {
		name string
		m    model.Metrics
		want string
	}{
		{"gauge is rounded to two decimals", model.Metrics{MType: model.Gauge, Value: ptr(3.14159)}, "3.14"},
		{"gauge with integer value", model.Metrics{MType: model.Gauge, Value: ptr(100.0)}, "100.00"},
		{"negative gauge", model.Metrics{MType: model.Gauge, Value: ptr(-0.5)}, "-0.50"},
		{"gauge without value", model.Metrics{MType: model.Gauge}, "0.00"},
		{"counter", model.Metrics{MType: model.Counter, Delta: ptr(int64(42))}, "42"},
		{"negative counter", model.Metrics{MType: model.Counter, Delta: ptr(int64(-7))}, "-7"},
		{"counter without delta", model.Metrics{MType: model.Counter}, "0"},
		{"unknown type has no value", model.Metrics{MType: "other"}, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, string(appendValue(nil, &tc.m)))
		})
	}
}

func TestAppendMainPage(t *testing.T) {
	t.Run("page without metrics", func(t *testing.T) {
		page := string(appendMainPage(nil, nil))
		assert.Contains(t, page, "Currently there are no metrics to display")
		assert.NotContains(t, page, "ID&nbsp;")
		assert.True(t, strings.HasPrefix(page, "<!DOCTYPE html>"))
		assert.True(t, strings.HasSuffix(page, "</html>\n"))
	})

	t.Run("page with metrics", func(t *testing.T) {
		page := string(appendMainPage(nil, []model.Metrics{
			{ID: "A", MType: model.Gauge, Value: ptr(1.5)},
			{ID: "B", MType: model.Counter, Delta: ptr(int64(2))},
		}))
		assert.Contains(t, page, "ID&nbsp;&nbsp;&nbsp;&nbsp;Value")
		assert.Contains(t, page, "<p>A&nbsp;&nbsp;&nbsp;&nbsp;1.50</p>")
		assert.Contains(t, page, "<p>B&nbsp;&nbsp;&nbsp;&nbsp;2</p>")
		assert.NotContains(t, page, "no metrics")
		assert.Less(t, strings.Index(page, "<p>A&"), strings.Index(page, "<p>B&"))
	})
}

// The page buffers are shared between requests, so concurrent requests must not see each other's data.
func TestMainPage_ConcurrentRequestsDoNotShareBuffers(t *testing.T) {
	s := repository.NewMemStorage()
	h := NewHandler(s, nil, nil, nil)

	const workers = 8
	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				w := httptest.NewRecorder()
				h.MainPage(w, httptest.NewRequest(http.MethodGet, "/", nil))
				body := w.Body.String()
				// A well-formed page: the head is first and the tail is last, even when the pool gives a used buffer.
				if !strings.HasPrefix(body, "<!DOCTYPE html>") || !strings.HasSuffix(body, "</html>\n") ||
					strings.Count(body, "<!DOCTYPE") != 1 {
					t.Errorf("malformed page: %q", body)
					return
				}
			}
		}()
	}

	// Meanwhile the number of metrics changes, so buffers of different sizes are reused.
	for i := 0; i < 100; i++ {
		v := float64(i)
		require.NoError(t, s.SaveMetrics(t.Context(), &model.Metrics{ID: fmt.Sprintf("m%d", i), MType: model.Gauge, Value: &v}))
	}
	wg.Wait()
}
