package handler

import (
	"net/url"
	"testing"

	models "github.com/paveltovchigrechko/metrics-service/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	s := models.NewStorage()
	h := NewHandler(s)

	assert.NotNil(t, h)
	assert.NotNil(t, h.storage)
}

func TestMainPage(t *testing.T) {

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
