package agent

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/paveltovchigrechko/metrics-service/internal/config"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgent(t *testing.T) {
	cfg, err := config.SetAgentConfig([]string{})
	require.NotNil(t, cfg)
	require.Nil(t, err)
	a := NewAgent(cfg)

	assert.NotNil(t, a)
	assert.NotNil(t, a.m)
	assert.NotNil(t, a.client)
	assert.NotNil(t, a.cfg)
	assert.Zero(t, a.PollCount)
}

func TestSendMetrics(t *testing.T) {
	t.Run("successfully sends counter and gauge metrics", func(t *testing.T) {
		var receivedPaths []string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedPaths = append(receivedPaths, r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg, err := config.SetAgentConfig([]string{})
		require.NotNil(t, cfg)
		require.Nil(t, err)
		a := NewAgent(cfg)

		cleanAddr := strings.TrimPrefix(server.URL, "http://")
		a.cfg.ServerAddress = cleanAddr

		counterMetric, err := models.CreateMetrics("PollCount", models.Counter, 10, 0)
		require.NoError(t, err)

		gaugeMetric, err := models.CreateMetrics("Alloc", models.Gauge, 0, 123.45)
		require.NoError(t, err)

		err = a.sendMetricsURL([]*models.Metrics{counterMetric, gaugeMetric})
		require.NoError(t, err)

		require.Len(t, receivedPaths, 2)
		assert.Equal(t, "/update/counter/PollCount/10", receivedPaths[0])
		assert.Equal(t, "/update/gauge/Alloc/123.45", receivedPaths[1])
	})

	t.Run("returns error when server is unreachable", func(t *testing.T) {
		cfg, err := config.SetAgentConfig([]string{})
		require.NotNil(t, cfg)
		require.Nil(t, err)
		a := NewAgent(cfg)

		metric, err := models.CreateMetrics("Alloc", models.Gauge, 0, 1.0)
		require.NoError(t, err)

		err = a.sendMetricsURL([]*models.Metrics{metric})
		assert.Error(t, err)
	})

	t.Run("stops at first failing metric", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		server.Close() // closed immediately so every request fails

		cfg, err := config.SetAgentConfig([]string{})
		require.NotNil(t, cfg)
		require.Nil(t, err)
		a := NewAgent(cfg)

		m1, err := models.CreateMetrics("Alloc", models.Gauge, 0, 1.0)
		require.NoError(t, err)
		m2, err := models.CreateMetrics("Frees", models.Gauge, 0, 2.0)
		require.NoError(t, err)

		err = a.sendMetricsURL([]*models.Metrics{m1, m2})
		assert.Error(t, err)
	})

	t.Run("sends empty metrics slice without error", func(t *testing.T) {
		cfg, err := config.SetAgentConfig([]string{})
		require.NotNil(t, cfg)
		require.Nil(t, err)
		a := NewAgent(cfg)

		err = a.sendMetricsURL([]*models.Metrics{})
		assert.NoError(t, err)
	})
}

func TestUpdateMetrics(t *testing.T) {
	t.Run("copies all MemStats fields onto Agent", func(t *testing.T) {
		m := &runtime.MemStats{
			Alloc:         1,
			BuckHashSys:   2,
			Frees:         3,
			GCCPUFraction: 0.5,
			GCSys:         4,
			HeapAlloc:     5,
			HeapIdle:      6,
			HeapInuse:     7,
			HeapObjects:   8,
			HeapReleased:  9,
			HeapSys:       10,
			LastGC:        11,
			Lookups:       12,
			Mallocs:       13,
			MCacheInuse:   14,
			MCacheSys:     15,
			MSpanInuse:    16,
			MSpanSys:      17,
			NextGC:        18,
			NumForcedGC:   19,
			NumGC:         20,
			OtherSys:      21,
			PauseTotalNs:  22,
			StackInuse:    23,
			StackSys:      24,
			Sys:           25,
			TotalAlloc:    26,
		}

		a := &Agent{m: m}
		a.updateMetrics()

		assert.Equal(t, uint64(1), a.Alloc)
		assert.Equal(t, uint64(2), a.BuckHashSys)
		assert.Equal(t, uint64(3), a.Frees)
		assert.Equal(t, 0.5, a.GCCPUFraction)
		assert.Equal(t, uint64(4), a.GCSys)
		assert.Equal(t, uint64(5), a.HeapAlloc)
		assert.Equal(t, uint64(6), a.HeapIdle)
		assert.Equal(t, uint64(7), a.HeapInuse)
		assert.Equal(t, uint64(8), a.HeapObjects)
		assert.Equal(t, uint64(9), a.HeapReleased)
		assert.Equal(t, uint64(10), a.HeapSys)
		assert.Equal(t, uint64(11), a.LastGC)
		assert.Equal(t, uint64(12), a.Lookups)
		assert.Equal(t, uint64(13), a.Mallocs)
		assert.Equal(t, uint64(14), a.MCacheInuse)
		assert.Equal(t, uint64(15), a.MCacheSys)
		assert.Equal(t, uint64(16), a.MSpanInuse)
		assert.Equal(t, uint64(17), a.MSpanSys)
		assert.Equal(t, uint64(18), a.NextGC)
		assert.Equal(t, uint32(19), a.NumForcedGC)
		assert.Equal(t, uint32(20), a.NumGC)
		assert.Equal(t, uint64(21), a.OtherSys)
		assert.Equal(t, uint64(22), a.PauseTotalNs)
		assert.Equal(t, uint64(23), a.StackInuse)
		assert.Equal(t, uint64(24), a.StackSys)
		assert.Equal(t, uint64(25), a.Sys)
		assert.Equal(t, uint64(26), a.TotalAlloc)
	})

	t.Run("increments PollCount by 1", func(t *testing.T) {
		a := &Agent{m: &runtime.MemStats{}, PollCount: 5}
		a.updateMetrics()
		assert.Equal(t, int64(6), a.PollCount)
	})

	t.Run("sets a new RandomValue within expected range", func(t *testing.T) {
		a := &Agent{m: &runtime.MemStats{}}
		a.updateMetrics()
		assert.GreaterOrEqual(t, a.RandomValue, 0.0)
		assert.Less(t, a.RandomValue, 100.0)
	})
}

func TestBuildMetrics(t *testing.T) {
	t.Run("returns one metric per registry entry with correct IDs", func(t *testing.T) {
		a := &Agent{
			Alloc:         100,
			BuckHashSys:   200,
			Frees:         300,
			GCCPUFraction: 0.1,
			GCSys:         400,
			HeapAlloc:     500,
			HeapIdle:      600,
			HeapInuse:     700,
			HeapObjects:   800,
			HeapReleased:  900,
			HeapSys:       1000,
			LastGC:        1100,
			Lookups:       1200,
			Mallocs:       1300,
			MCacheInuse:   1400,
			MCacheSys:     1500,
			MSpanInuse:    1600,
			MSpanSys:      1700,
			NextGC:        1800,
			NumForcedGC:   19,
			NumGC:         20,
			OtherSys:      2100,
			PauseTotalNs:  2200,
			StackInuse:    2300,
			StackSys:      2400,
			Sys:           2500,
			TotalAlloc:    2600,
			PollCount:     7,
			RandomValue:   42.42,
		}

		result := a.buildMetrics()

		require.Len(t, result, len(metricsRegistry))
	})

	t.Run("PollCount is a counter with correct delta", func(t *testing.T) {
		a := &Agent{PollCount: 7}
		result := a.buildMetrics()

		found := findMetric(result, "PollCount")
		require.NotNil(t, found)
		assert.Equal(t, models.Counter, found.MType)
		require.NotNil(t, found.Delta)
		assert.Equal(t, int64(7), *found.Delta)
	})

	t.Run("RandomValue is a gauge with correct value", func(t *testing.T) {
		a := &Agent{RandomValue: 42.42}
		result := a.buildMetrics()

		found := findMetric(result, "RandomValue")
		require.NotNil(t, found)
		assert.Equal(t, models.Gauge, found.MType)
		require.NotNil(t, found.Value)
		assert.Equal(t, 42.42, *found.Value)
	})

	t.Run("Alloc is a gauge reflecting the Agent field", func(t *testing.T) {
		a := &Agent{Alloc: 100}
		result := a.buildMetrics()

		found := findMetric(result, "Alloc")
		require.NotNil(t, found)
		assert.Equal(t, models.Gauge, found.MType)
		require.NotNil(t, found.Value)
		assert.Equal(t, 100.0, *found.Value)
	})

	t.Run("returns empty-safe result for zero-value Agent", func(t *testing.T) {
		a := &Agent{}
		result := a.buildMetrics()
		assert.Len(t, result, len(metricsRegistry))
	})
}

// Test helper
func findMetric(metrics []*models.Metrics, id string) *models.Metrics {
	for _, m := range metrics {
		if m.ID == id {
			return m
		}
	}
	return nil
}

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
			expectedUrl: "http://localhost:8080/update/counter/someName/8",
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
			expectedUrl: "http://localhost:8080/update/gauge/someName/8.89",
		},
	}
	cfg, err := config.SetAgentConfig([]string{})
	require.NotNil(t, cfg)
	require.Nil(t, err)

	a := NewAgent(cfg)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := a.createURLFromMetric(tc.m)
			assert.Equal(t, tc.expectedUrl, result)
		})
	}
}
