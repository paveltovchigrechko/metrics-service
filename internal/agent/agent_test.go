package agent

import (
	"log"
	"testing"

	models "github.com/paveltovchigrechko/metrics-service/internal/model"
)

func TestCreateURLFromMetric(t *testing.T) {
	testCases := []struct {
		name        string
		m           *models.Metrics
		expectedUrl string
	}{
		{
			name: "Counter metrics",
			m: func() *models.Metrics {
				delta := int64(8)
				return &models.Metrics{
					ID:    "someName",
					MType: models.Counter,
					Delta: &delta,
				}
			}(),
			expectedUrl: host + "/update/counter/someName/8",
		},
		{
			name: "Gauge metrics",
			m: func() *models.Metrics {
				value := float64(8.8893)
				return &models.Metrics{
					ID:    "someName",
					MType: models.Gauge,
					Value: &value,
				}
			}(),
			expectedUrl: host + "/update/gauge/someName/8.89",
		},
	}

	for _, test := range testCases {
		result := createURLFromMetric(test.m)
		log.Print(result)
		if result != test.expectedUrl {
			t.Error()
		}
	}
}
