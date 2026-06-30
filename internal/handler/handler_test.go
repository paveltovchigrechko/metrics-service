package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	s := models.NewStorage()
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
			headers:  map[string]string{}, // Empty headers
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

			s := models.NewStorage()
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

func TestProcessMetrics(t *testing.T) {
	s := models.NewStorage()
	h := NewHandler(s)

	counterUrl, _ := url.Parse("http://some-address.domain/update/counter/counter-name/77")
	gaugeUrl, _ := url.Parse("http://some-address.domain/update/gauge/gauge-name/77.76")
	testCases := []struct {
		name string
		url  *url.URL
	}{
		{
			"URL for counter metrics",
			counterUrl,
		},
		{
			"URL for gauge metrics",
			gaugeUrl,
		},
	}

	expectedMEtrics := []*models.Metrics{
		{
			ID:    "counter-name",
			MType: "counter",
			Delta: &[]int64{77}[0],
		},
		{
			ID:    "gauge-name",
			MType: "gauge",
			Value: &[]float64{77.76}[0],
		},
	}

	for i, test := range testCases {
		h.processMetrics(test.url)

		assert.Equal(t, expectedMEtrics[i].ID, h.storage.Metrics[expectedMEtrics[i].ID].ID)
		assert.Equal(t, expectedMEtrics[i].MType, h.storage.Metrics[expectedMEtrics[i].ID].MType)

		switch expectedMEtrics[i].MType {
		case models.Counter:
			assert.Equal(t, expectedMEtrics[0].Delta, h.storage.Metrics[expectedMEtrics[0].ID].Delta)
			assert.Nil(t, h.storage.Metrics[expectedMEtrics[0].ID].Value)
		case models.Gauge:
			assert.Equal(t, expectedMEtrics[1].Value, h.storage.Metrics[expectedMEtrics[1].ID].Value)
			assert.Nil(t, h.storage.Metrics[expectedMEtrics[1].ID].Delta)
		}
	}
}

func TestParseMetrics(t *testing.T) {
	// Process errors?
	counterUrl, _ := url.Parse("http://some-address.domain/update/counter/counter-name/77")
	gaugeUrl, _ := url.Parse("http://some-address.domain/update/gauge/gauge-name/77.76")

	testCases := []struct {
		name    string
		url     *url.URL
		metrics *models.Metrics
	}{
		{
			name: "URL for counter metrics",
			url:  counterUrl,
			metrics: &models.Metrics{
				ID:    "counter-name",
				MType: "counter",
				Delta: &[]int64{77}[0],
			},
		},
		{
			name: "URL for gauge metrics",
			url:  gaugeUrl,
			metrics: &models.Metrics{
				ID:    "gauge-name",
				MType: "gauge",
				Value: &[]float64{77.76}[0],
			},
		},
	}

	for _, test := range testCases {
		result := parseMetrics(test.url)
		assert.Equal(t, test.metrics.ID, result.ID)
		assert.Equal(t, test.metrics.MType, result.MType)

		switch test.metrics.MType {
		case models.Counter:
			assert.Equal(t, *test.metrics.Delta, *result.Delta)
			assert.Nil(t, result.Value)
		case models.Gauge:
			assert.Equal(t, *test.metrics.Value, *result.Value)
			assert.Nil(t, result.Delta)
		}
	}
}
