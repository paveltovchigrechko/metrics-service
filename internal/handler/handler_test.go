package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockStorage struct {
	OnGetMetrics  func(string) (*models.Metrics, error)
	OnSaveMetrics func(*models.Metrics) error
	OnListMetrics func(io.Writer)
}

func (m *MockStorage) GetMetrics(name string) (*models.Metrics, error) { return m.OnGetMetrics(name) }
func (m *MockStorage) SaveMetrics(mt *models.Metrics) error            { return m.OnSaveMetrics(mt) }
func (m *MockStorage) ListMetrics(w io.Writer)                         { m.OnListMetrics(w) }

func newRequestWithChiParams(method, target string, params map[string]string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	chiCtx := chi.NewRouteContext()
	for k, v := range params {
		chiCtx.URLParams.Add(k, v)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
}

func TestNewHandler(t *testing.T) {
	s := models.NewMemStorage()
	h := NewHandler(s)

	assert.NotNil(t, h)
	assert.NotNil(t, h.storage)
}

func TestPostMetrics(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		target   string
		headers  map[string]string
		wantCode int
		wantType string
	}{
		{
			name:     "positive test #1 (valid gauge)",
			method:   http.MethodPost,
			target:   "/update/gauge/Alloc/1024.50",
			headers:  map[string]string{"Content-Type": "text/plain"},
			wantCode: http.StatusOK,
			wantType: "",
		},
		{
			name:     "negative test - unsupported endpoint",
			method:   http.MethodPost,
			target:   "/server-updates/gauge/Alloc/1024.50",
			headers:  map[string]string{"Content-Type": "text/plain"},
			wantCode: http.StatusBadRequest,
			wantType: "",
		},
		{
			name:     "negative test - unsupported method",
			method:   http.MethodGet,
			target:   "/update/gauge/Alloc/1024.50",
			headers:  map[string]string{"Content-Type": "text/plain"},
			wantCode: http.StatusBadRequest,
			wantType: "",
		},
		{
			name:     "positive test - missing content type",
			method:   http.MethodPost,
			target:   "/update/gauge/Alloc/1024.50",
			headers:  map[string]string{},
			wantCode: http.StatusOK,
			wantType: "",
		},
		{
			name:     "negative test - 400 no metric",
			method:   http.MethodPost,
			target:   "/update/gauge/",
			headers:  map[string]string{"Content-Type": "text/plain"},
			wantCode: http.StatusBadRequest,
			wantType: "",
		},
		{
			name:     "negative test - 400 invalid metric value",
			method:   http.MethodPost,
			target:   "/update/gauge/Alloc/abc",
			headers:  map[string]string{"Content-Type": "text/plain"},
			wantCode: http.StatusBadRequest,
			wantType: "",
		},
		{
			name:     "negative test - 400 invalid metric type",
			method:   http.MethodPost,
			target:   "/update/unknown-type/Alloc/100",
			headers:  map[string]string{"Content-Type": "text/plain"},
			wantCode: http.StatusBadRequest,
			wantType: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, nil)
			for key, value := range test.headers {
				request.Header.Set(key, value)
			}

			s := models.NewMemStorage()
			h := NewHandler(s)
			r := chi.NewRouter()

			r.Method(http.MethodPost, "/update/{metricsType}/{metricsName}/{metricsValue}", http.HandlerFunc(h.PostMetrics))
			r.NotFound(h.PostMetrics)
			r.MethodNotAllowed(h.PostMetrics)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.wantCode, res.StatusCode)
			if test.wantCode == http.StatusOK {
				assert.Equal(t, test.wantType, res.Header.Get("Content-Type"))
			}
		})
	}
}

// func TestProcessMetrics(t *testing.T) {
// 	testCases := []struct {
// 		name         string
// 		req          *http.Request
// 		expectedName string
// 		expectedType string
// 		expectErr    bool
// 		expectSaved  bool
// 		verify       func(t *testing.T, m *models.Metrics)
// 	}{
// 		{
// 			name: "processes valid counter metric successfully",
// 			req: newRequestWithChiParams(http.MethodPost, "/update/counter/counter-name/77", map[string]string{
// 				"metricsType":  "counter",
// 				"metricsName":  "counter-name",
// 				"metricsValue": "77",
// 			}),
// 			expectedName: "counter-name",
// 			expectedType: "counter",
// 			expectErr:    false,
// 			expectSaved:  true,
// 			verify: func(t *testing.T, m *models.Metrics) {
// 				assert.Equal(t, "counter", m.MType)
// 				assert.Equal(t, int64(77), *m.Delta)
// 				assert.Nil(t, m.Value)
// 			},
// 		},
// 		{
// 			name: "processes valid gauge metric successfully",
// 			req: newRequestWithChiParams(http.MethodPost, "/update/gauge/gauge-name/77.76", map[string]string{
// 				"metricsType":  "gauge",
// 				"metricsName":  "gauge-name",
// 				"metricsValue": "77.76",
// 			}),
// 			expectedName: "gauge-name",
// 			expectedType: "gauge",
// 			expectErr:    false,
// 			expectSaved:  true,
// 			verify: func(t *testing.T, m *models.Metrics) {
// 				assert.Equal(t, "gauge", m.MType)
// 				assert.Equal(t, 77.76, *m.Value)
// 				assert.Nil(t, m.Delta)
// 			},
// 		},
// 		{
// 			name: "returns error on parsing failure",
// 			req: newRequestWithChiParams(http.MethodPost, "/update/counter/bad-counter/abc", map[string]string{
// 				"metricsType":  "counter",
// 				"metricsName":  "bad-counter",
// 				"metricsValue": "abc", // Invalid digits for integer
// 			}),
// 			expectedName: "bad-counter",
// 			expectedType: "counter",
// 			expectErr:    true,
// 			expectSaved:  false,
// 			verify:       nil,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			s := models.NewMemStorage()
// 			h := NewHandler(s)

// 			err := h.processMetrics(tc.req)

// 			if tc.expectErr {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}

// 			metric, getErr := h.storage.GetMetrics(tc.expectedName, tc.expectedType)
// 			if tc.expectSaved {
// 				require.NoError(t, getErr)
// 				require.NotNil(t, metric)
// 				tc.verify(t, metric)
// 			} else {
// 				assert.Error(t, getErr, "expected metric to not exist in storage")
// 				assert.Nil(t, metric)
// 			}
// 		})
// 	}
// }

func TestParseMetrics(t *testing.T) {
	testCases := []struct {
		name      string
		req       *http.Request
		expectErr bool
		verify    func(t *testing.T, result *models.Metrics)
	}{
		{
			name: "successfully parses counter metrics",
			req: newRequestWithChiParams(http.MethodPost, "/update/counter/counter-name/77", map[string]string{
				"metricsType":  "counter",
				"metricsName":  "counter-name",
				"metricsValue": "77",
			}),
			expectErr: false,
			verify: func(t *testing.T, result *models.Metrics) {
				require.NotNil(t, result)
				assert.Equal(t, "counter:counter-name", result.ID)
				assert.Equal(t, "counter", result.MType)
				assert.Equal(t, int64(77), *result.Delta)
				assert.Nil(t, result.Value)
			},
		},
		{
			name: "successfully parses gauge metrics",
			req: newRequestWithChiParams(http.MethodPost, "/update/gauge/gauge-name/77.76", map[string]string{
				"metricsType":  "gauge",
				"metricsName":  "gauge-name",
				"metricsValue": "77.76",
			}),
			expectErr: false,
			verify: func(t *testing.T, result *models.Metrics) {
				require.NotNil(t, result)
				assert.Equal(t, "gauge:gauge-name", result.ID)
				assert.Equal(t, "gauge", result.MType)
				assert.Equal(t, 77.76, *result.Value)
				assert.Nil(t, result.Delta)
			},
		},
		{
			name: "returns error for invalid counter integer format",
			req: newRequestWithChiParams(http.MethodPost, "/update/counter/counter-name/12.34", map[string]string{
				"metricsType":  "counter",
				"metricsName":  "counter-name",
				"metricsValue": "12.34", // Float value into counter
			}),
			expectErr: true,
			verify: func(t *testing.T, result *models.Metrics) {
				assert.Nil(t, result)
			},
		},
		{
			name: "returns error for invalid gauge float format",
			req: newRequestWithChiParams(http.MethodPost, "/update/gauge/gauge-name/not-a-number", map[string]string{
				"metricsType":  "gauge",
				"metricsName":  "gauge-name",
				"metricsValue": "not-a-number",
			}),
			expectErr: true,
			verify: func(t *testing.T, result *models.Metrics) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseMetrics(tc.req)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			tc.verify(t, result)
		})
	}
}
