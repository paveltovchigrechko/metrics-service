package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/paveltovchigrechko/metrics-service/internal/common"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Pointer helpers for types
func ptr[T any](v T) *T {
	return &v
}

// MockStorage implements models.Storage explicitly.
type MockStorage struct {
	OnGetMetrics     func(string, string) (*models.Metrics, error)
	OnListMetrics    func(io.Writer)
	OnSaveMetrics    func(*models.Metrics) error
	OnRestoreMetrics func(*models.Metrics) error
	OnGetAllMetrics  func() []models.Metrics
}

func (m *MockStorage) GetMetrics(name, mtype string) (*models.Metrics, error) {
	if m.OnGetMetrics != nil {
		return m.OnGetMetrics(name, mtype)
	}
	return nil, nil
}

func (m *MockStorage) ListMetrics(w io.Writer) {
	if m.OnListMetrics != nil {
		m.OnListMetrics(w)
	}
}

func (m *MockStorage) SaveMetrics(mt *models.Metrics) error {
	if m.OnSaveMetrics != nil {
		return m.OnSaveMetrics(mt)
	}
	return nil
}

func (m *MockStorage) RestoreMetrics(mt *models.Metrics) error {
	if m.OnRestoreMetrics != nil {
		return m.OnRestoreMetrics(mt)
	}
	return nil
}

func (m *MockStorage) GetAllMetrics() []models.Metrics {
	if m.OnGetAllMetrics != nil {
		return m.OnGetAllMetrics()
	}
	return nil
}

func newRequestWithChiParams(method, target string, params map[string]string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	chiCtx := chi.NewRouteContext()
	for k, v := range params {
		chiCtx.URLParams.Add(k, v)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
}

// ==========================================
// CORE HANDLER UNIT TESTS
// ==========================================

func TestNewHandler(t *testing.T) {
	s := models.NewMemStorage()
	h := NewHandler(s, nil)

	assert.NotNil(t, h)
	assert.NotNil(t, h.storage)
}

func TestPostMetrics(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		target    string
		headers   map[string]string
		updateErr error
		wantCode  int
		wantType  string
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
			name:     "positive test - missing content type defaults to text/plain",
			method:   http.MethodPost,
			target:   "/update/gauge/Alloc/1024.50",
			headers:  map[string]string{},
			wantCode: http.StatusOK,
			wantType: "",
		},
		{
			name:     "negative test - 400 no metric name",
			method:   http.MethodPost,
			target:   "/update/gauge/",
			headers:  map[string]string{"Content-Type": "text/plain"},
			wantCode: http.StatusBadRequest,
			wantType: "",
		},
		{
			name:     "negative test - 400 invalid metric value formatting",
			method:   http.MethodPost,
			target:   "/update/gauge/Alloc/abc",
			headers:  map[string]string{"Content-Type": "text/plain"},
			wantCode: http.StatusBadRequest,
			wantType: "",
		},
		{
			name:     "negative test - 400 invalid metric type value",
			method:   http.MethodPost,
			target:   "/update/unknown-type/Alloc/100",
			headers:  map[string]string{"Content-Type": "text/plain"},
			wantCode: http.StatusBadRequest,
			wantType: "",
		},
		{
			name:      "negative test - 500 when synchronous afterSuccessfulUpdate fails",
			method:    http.MethodPost,
			target:    "/update/gauge/Alloc/100",
			headers:   map[string]string{"Content-Type": "text/plain"},
			updateErr: errors.New("failed to flush data"),
			wantCode:  http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, nil)
			for key, value := range test.headers {
				request.Header.Set(key, value)
			}

			s := models.NewMemStorage()
			var updateFunc func() error
			if test.updateErr != nil {
				updateFunc = func() error { return test.updateErr }
			}
			h := NewHandler(s, updateFunc)
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

func TestMainPage(t *testing.T) {
	t.Run("successfully renders html list", func(t *testing.T) {
		mockStorage := &MockStorage{
			OnGetAllMetrics: func() []models.Metrics {
				return []models.Metrics{
					{
						ID:    "Alloc",
						MType: models.Gauge,
						Value: ptr(100.00),
					},
					{
						ID:    "PollCount",
						MType: models.Counter,
						Delta: ptr(int64(5)),
					},
				}
			},
		}
		h := NewHandler(mockStorage, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		h.MainPage(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/html")

		bodyStr := w.Body.String()
		assert.Contains(t, bodyStr, "Welcome to the Metrics service")
		assert.Contains(t, bodyStr, "ID&nbsp;&nbsp;&nbsp;&nbsp;Value")

		assert.Contains(t, bodyStr, "Alloc")
		assert.Contains(t, bodyStr, "100.00")
		assert.Contains(t, bodyStr, "PollCount")
		assert.Contains(t, bodyStr, "5")
	})

	t.Run("renders empty state message when no metrics exist", func(t *testing.T) {
		mockStorage := &MockStorage{
			OnGetAllMetrics: func() []models.Metrics {
				return []models.Metrics{}
			},
		}
		h := NewHandler(mockStorage, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		h.MainPage(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Currently there are no metrics to display")
	})
}

func TestMetricsValue(t *testing.T) {
	tests := []struct {
		name         string
		params       map[string]string
		mockStorage  func() models.Storage
		expectedCode int
		expectedBody string
	}{
		{
			name: "successful counter extraction",
			params: map[string]string{
				"metricsType": "counter",
				"metricsName": "PollCount",
			},
			mockStorage: func() models.Storage {
				m := &MockStorage{}
				m.OnGetMetrics = func(name, mtype string) (*models.Metrics, error) {
					return &models.Metrics{ID: "PollCount", MType: "counter", Delta: ptr(int64(45))}, nil
				}
				return m
			},
			expectedCode: http.StatusOK,
			expectedBody: "45",
		},
		{
			name: "successful gauge extraction",
			params: map[string]string{
				"metricsType": "gauge",
				"metricsName": "Alloc",
			},
			mockStorage: func() models.Storage {
				m := &MockStorage{}
				m.OnGetMetrics = func(name, mtype string) (*models.Metrics, error) {
					return &models.Metrics{ID: "Alloc", MType: "gauge", Value: ptr(1234.56)}, nil
				}
				return m
			},
			expectedCode: http.StatusOK,
			expectedBody: "1234.56",
		},
		{
			name: "returns 404 when metric not found in storage",
			params: map[string]string{
				"metricsType": "gauge",
				"metricsName": "NonExistent",
			},
			mockStorage: func() models.Storage {
				m := &MockStorage{}
				m.OnGetMetrics = func(name, mtype string) (*models.Metrics, error) {
					return nil, errors.New("not found")
				}
				return m
			},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.mockStorage(), nil)
			req := newRequestWithChiParams(http.MethodGet, "/", tt.params)
			w := httptest.NewRecorder()

			h.MetricsValue(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedCode == http.StatusOK {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestUpdateEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		updateErr   error
		wantStatus  int
	}{
		{
			name:        "valid counter update",
			contentType: "application/json",
			body:        `{"id": "PollCount", "type": "counter", "delta": 10}`,
			wantStatus:  http.StatusOK,
		},
		{
			name:        "valid gauge update",
			contentType: "application/json",
			body:        `{"id": "Alloc", "type": "gauge", "value": 154.32}`,
			wantStatus:  http.StatusOK,
		},
		{
			name:        "invalid content type",
			contentType: "text/plain",
			body:        `{"id": "Alloc", "type": "gauge", "value": 154.32}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "missing delta for counter",
			contentType: "application/json",
			body:        `{"id": "PollCount", "type": "counter"}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "malformed json payload",
			contentType: "application/json",
			body:        `{"id": "PollCount", "type": "counter",`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "failed synchronous flush",
			contentType: "application/json",
			body:        `{"id": "Alloc", "type": "gauge", "value": 154.32}`,
			updateErr:   errors.New("disk full"),
			wantStatus:  http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := models.NewMemStorage()
			var updateFunc func() error
			if tt.updateErr != nil {
				updateFunc = func() error { return tt.updateErr }
			}
			h := NewHandler(s, updateFunc)

			req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()
			h.UpdateEndpoint(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestValueEndpoint(t *testing.T) {
	tests := []struct {
		name         string
		contentType  string
		body         string
		setupStorage func() models.Storage
		wantStatus   int
		wantInBody   string
	}{
		{
			name:        "successfully get counter value",
			contentType: "application/json",
			body:        `{"id": "PollCount", "type": "counter"}`,
			setupStorage: func() models.Storage {
				m := &MockStorage{}
				m.OnGetMetrics = func(name, mtype string) (*models.Metrics, error) {
					return &models.Metrics{ID: "PollCount", MType: "counter", Delta: ptr(int64(42))}, nil
				}
				return m
			},
			wantStatus: http.StatusOK,
			wantInBody: `"delta":42`,
		},
		{
			name:        "successfully get gauge value",
			contentType: "application/json",
			body:        `{"id": "Alloc", "type": "gauge"}`,
			setupStorage: func() models.Storage {
				m := &MockStorage{}
				m.OnGetMetrics = func(name, mtype string) (*models.Metrics, error) {
					return &models.Metrics{ID: "Alloc", MType: "gauge", Value: ptr(12.34)}, nil
				}
				return m
			},
			wantStatus: http.StatusOK,
			wantInBody: `"value":12.34`,
		},
		{
			name:        "metric not found",
			contentType: "application/json",
			body:        `{"id": "Missing", "type": "gauge"}`,
			setupStorage: func() models.Storage {
				m := &MockStorage{}
				m.OnGetMetrics = func(name, mtype string) (*models.Metrics, error) {
					return nil, errors.New("not found")
				}
				return m
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:        "incorrect content type",
			contentType: "text/plain",
			body:        `{"id": "Alloc", "type": "gauge"}`,
			setupStorage: func() models.Storage {
				return &MockStorage{}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:        "type mismatch in found metric",
			contentType: "application/json",
			body:        `{"id": "Alloc", "type": "counter"}`, // requesting counter
			setupStorage: func() models.Storage {
				m := &MockStorage{}
				m.OnGetMetrics = func(name, mtype string) (*models.Metrics, error) {
					return &models.Metrics{ID: "Alloc", MType: "gauge", Value: ptr(12.34)}, nil // returns gauge
				}
				return m
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.setupStorage(), nil)

			req := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()
			h.ValueEndpoint(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				assert.Contains(t, w.Body.String(), tt.wantInBody)
			}
		})
	}
}

func TestProcessMetrics(t *testing.T) {
	testCases := []struct {
		name         string
		req          *http.Request
		expectedName string
		expectedType string
		expectErr    bool
		expectSaved  bool
		verify       func(t *testing.T, m *models.Metrics)
	}{
		{
			name: "processes valid counter metric successfully",
			req: newRequestWithChiParams(http.MethodPost, "/update/counter/counter-name/77", map[string]string{
				"metricsType":  "counter",
				"metricsName":  "counter-name",
				"metricsValue": "77",
			}),
			expectedName: "counter-name",
			expectedType: "counter",
			expectErr:    false,
			expectSaved:  true,
			verify: func(t *testing.T, m *models.Metrics) {
				assert.Equal(t, "counter", m.MType)
				assert.Equal(t, int64(77), *m.Delta)
				assert.Nil(t, m.Value)
			},
		},
		{
			name: "processes valid gauge metric successfully",
			req: newRequestWithChiParams(http.MethodPost, "/update/gauge/gauge-name/77.76", map[string]string{
				"metricsType":  "gauge",
				"metricsName":  "gauge-name",
				"metricsValue": "77.76",
			}),
			expectedName: "gauge-name",
			expectedType: "gauge",
			expectErr:    false,
			expectSaved:  true,
			verify: func(t *testing.T, m *models.Metrics) {
				assert.Equal(t, "gauge", m.MType)
				assert.Equal(t, 77.76, *m.Value)
				assert.Nil(t, m.Delta)
			},
		},
		{
			name: "returns error on parsing failure",
			req: newRequestWithChiParams(http.MethodPost, "/update/counter/bad-counter/abc", map[string]string{
				"metricsType":  "counter",
				"metricsName":  "bad-counter",
				"metricsValue": "abc",
			}),
			expectedName: "bad-counter",
			expectedType: "counter",
			expectErr:    true,
			expectSaved:  false,
			verify:       nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := models.NewMemStorage()
			h := NewHandler(s, nil)

			err := h.processMetrics(tc.req)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			metric, getErr := h.storage.GetMetrics(tc.expectedName, tc.expectedType)
			if tc.expectSaved {
				require.NoError(t, getErr)
				require.NotNil(t, metric)
				tc.verify(t, metric)
			} else {
				assert.Error(t, getErr, "expected metric to not exist in storage")
				assert.Nil(t, metric)
			}
		})
	}
}

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
				assert.Equal(t, "counter-name", result.ID)
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
				assert.Equal(t, "gauge-name", result.ID)
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
				"metricsValue": "12.34",
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

func TestParseUpdateMetrics(t *testing.T) {
	tests := []struct {
		name      string
		input     *common.Metrics
		expectErr bool
		verify    func(t *testing.T, m *models.Metrics)
	}{
		{
			name: "valid counter translation",
			input: &common.Metrics{
				ID:    "PollCount",
				MType: "counter",
				Delta: ptr(int64(5)),
			},
			expectErr: false,
			verify: func(t *testing.T, m *models.Metrics) {
				require.NotNil(t, m)
				assert.Equal(t, "PollCount", m.ID)
				assert.Equal(t, "counter", m.MType)
				assert.Equal(t, int64(5), *m.Delta)
			},
		},
		{
			name: "valid gauge translation",
			input: &common.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: ptr(743.21),
			},
			expectErr: false,
			verify: func(t *testing.T, m *models.Metrics) {
				require.NotNil(t, m)
				assert.Equal(t, "Alloc", m.ID)
				assert.Equal(t, "gauge", m.MType)
				assert.Equal(t, 743.21, *m.Value)
			},
		},
		{
			name: "empty metric id",
			input: &common.Metrics{
				MType: "counter",
				Delta: ptr(int64(5)),
			},
			expectErr: true,
		},
		{
			name: "missing counter delta",
			input: &common.Metrics{
				ID:    "PollCount",
				MType: "counter",
			},
			expectErr: true,
		},
		{
			name: "missing gauge value",
			input: &common.Metrics{
				ID:    "Alloc",
				MType: "gauge",
			},
			expectErr: true,
		},
		{
			name: "unknown metric type",
			input: &common.Metrics{
				ID:    "Alloc",
				MType: "invalid-type",
				Value: ptr(10.0),
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parseUpdateMetrics(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.verify(t, res)
			}
		})
	}
}

func TestParseValueMetrics(t *testing.T) {
	tests := []struct {
		name      string
		input     *common.Metrics
		expectErr bool
		wantID    string
		wantType  string
	}{
		{
			name:      "valid counter template retrieval",
			input:     &common.Metrics{ID: "PollCount", MType: "counter"},
			expectErr: false,
			wantID:    "PollCount",
			wantType:  "counter",
		},
		{
			name:      "valid gauge template retrieval",
			input:     &common.Metrics{ID: "Alloc", MType: "gauge"},
			expectErr: false,
			wantID:    "Alloc",
			wantType:  "gauge",
		},
		{
			name:      "missing metric id",
			input:     &common.Metrics{MType: "counter"},
			expectErr: true,
		},
		{
			name:      "unknown metrics type value request",
			input:     &common.Metrics{ID: "Alloc", MType: "unknown"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, mtype, err := parseValueMetrics(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, id)
				assert.Equal(t, tt.wantType, mtype)
			}
		})
	}
}

func TestDecodeJSONMetrics(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		expectErr bool
		verify    func(t *testing.T, m *common.Metrics)
	}{
		{
			name:      "decode completely valid gauge struct",
			body:      `{"id":"Alloc","type":"gauge","value":54.12}`,
			expectErr: false,
			verify: func(t *testing.T, m *common.Metrics) {
				require.NotNil(t, m)
				assert.Equal(t, "Alloc", m.ID)
				assert.Equal(t, "gauge", m.MType)
				assert.Equal(t, 54.12, *m.Value)
			},
		},
		{
			name:      "reject unknown fields due to strict decoding settings",
			body:      `{"id":"Alloc","type":"gauge","value":54.12,"unexpected_field":"hello"}`,
			expectErr: true,
		},
		{
			name:      "malformed partial json syntax error",
			body:      `{"id":"Alloc",`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			res, err := decodeJSONMetrics(req)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.verify(t, res)
			}
		})
	}
}
