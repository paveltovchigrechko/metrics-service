package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/paveltovchigrechko/metrics-service/internal/config"
	"github.com/paveltovchigrechko/metrics-service/internal/hashing"
	"github.com/paveltovchigrechko/metrics-service/internal/model"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

const (
	textPlain       = "text/plain"
	applicationJSON = "application/json"
	gzipEncoding    = "gzip"
	reqTimeout      = 5 * time.Second
)

type Agent struct {
	m      *runtime.MemStats
	client *resty.Client
	cfg    *config.AgentConfig
	mu     sync.Mutex

	// Runtime metrics
	Alloc         uint64
	BuckHashSys   uint64
	Frees         uint64
	GCCPUFraction float64
	GCSys         uint64
	HeapAlloc     uint64
	HeapIdle      uint64
	HeapInuse     uint64
	HeapObjects   uint64
	HeapReleased  uint64
	HeapSys       uint64
	LastGC        uint64
	Lookups       uint64
	Mallocs       uint64
	MCacheInuse   uint64
	MCacheSys     uint64
	MSpanInuse    uint64
	MSpanSys      uint64
	NextGC        uint64
	NumForcedGC   uint32
	NumGC         uint32
	OtherSys      uint64
	PauseTotalNs  uint64
	StackInuse    uint64
	StackSys      uint64
	Sys           uint64
	TotalAlloc    uint64

	PollCount   int64
	RandomValue float64

	// Metrics from gopsutil
	TotalMemory     float64
	FreeMemory      float64
	CPUUtilizations []float64
}

func NewAgent(cfg *config.AgentConfig) *Agent {
	m := new(runtime.MemStats)

	c := resty.New().
		SetTimeout(reqTimeout).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			if err != nil {
				return true
			}

			return false
		})

	return &Agent{
		m:           m,
		client:      c,
		cfg:         cfg,
		PollCount:   0,
		RandomValue: calcNewRandomValue(),
	}
}

func (a *Agent) Run(ctx context.Context) {
	metricsChan := make(chan []model.Metrics, a.cfg.RateLimit*2)

	for i := 0; i < a.cfg.RateLimit; i++ {
		// RateLimit goroutines that read channel with metrics and send them.
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case metrics, ok := <-metricsChan:
					if !ok {
						return
					}
					if err := a.sendMetricsJSON(metrics, true); err != nil {
						log.Printf("failed to send metrics: %v", err)
					}
				}
			}
		}()
	}

	// Separate goroutines for runtime and PSU metrics and reporting.
	go a.collectRuntimeMetrics(ctx, a.cfg.PollInterval)
	go a.collectPSUMetrics(ctx, a.cfg.PollInterval)
	go a.startReporting(ctx, a.cfg.ReportInterval, metricsChan)
	<-ctx.Done()
}

func (a *Agent) collectRuntimeMetrics(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runtime.ReadMemStats(a.m)
			a.updateMetrics()
		}
	}
}

func (a *Agent) collectPSUMetrics(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.updatePSUMetrics(); err != nil {
				log.Print("failed to fetch PSU metrics: ", err)
			}
		}
	}
}

func (a *Agent) startReporting(ctx context.Context, interval time.Duration, metricsChan chan<- []model.Metrics) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics := a.buildMetrics()

			select {
			case metricsChan <- metrics:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (a *Agent) updateMetrics() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Use reflect to copy values?
	a.Alloc = a.m.Alloc
	a.BuckHashSys = a.m.BuckHashSys
	a.Frees = a.m.Frees
	a.GCCPUFraction = a.m.GCCPUFraction
	a.GCSys = a.m.GCSys
	a.HeapAlloc = a.m.HeapAlloc
	a.HeapIdle = a.m.HeapIdle
	a.HeapInuse = a.m.HeapInuse
	a.HeapObjects = a.m.HeapObjects
	a.HeapReleased = a.m.HeapReleased
	a.HeapSys = a.m.HeapSys
	a.LastGC = a.m.LastGC
	a.Lookups = a.m.Lookups
	a.Mallocs = a.m.Mallocs
	a.MCacheInuse = a.m.MCacheInuse
	a.MCacheSys = a.m.MCacheSys
	a.MSpanInuse = a.m.MSpanInuse
	a.MSpanSys = a.m.MSpanSys
	a.NextGC = a.m.NextGC
	a.NumForcedGC = a.m.NumForcedGC
	a.NumGC = a.m.NumGC
	a.OtherSys = a.m.OtherSys
	a.PauseTotalNs = a.m.PauseTotalNs
	a.StackInuse = a.m.StackInuse
	a.StackSys = a.m.StackSys
	a.Sys = a.m.Sys
	a.TotalAlloc = a.m.TotalAlloc

	a.PollCount++
	a.RandomValue = calcNewRandomValue()
}

func (a *Agent) updatePSUMetrics() error {
	v, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.TotalMemory = float64(v.Total)
	a.FreeMemory = float64(v.Free)
	cpuPercentages, err := cpu.Percent(0, true)
	if err != nil {
		return err
	}

	a.CPUUtilizations = cpuPercentages

	return nil
}

type MetricDescriptor struct {
	Get func(*Agent) *models.Metrics
}

var metricsRegistry = []MetricDescriptor{
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.Alloc)
			m, _ := models.CreateMetrics("Alloc", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.BuckHashSys)
			m, _ := models.CreateMetrics("BuckHashSys", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.Frees)
			m, _ := models.CreateMetrics("Frees", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := a.GCCPUFraction
			m, _ := models.CreateMetrics("GCCPUFraction", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.GCSys)
			m, _ := models.CreateMetrics("GCSys", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.HeapAlloc)
			m, _ := models.CreateMetrics("HeapAlloc", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.HeapIdle)
			m, _ := models.CreateMetrics("HeapIdle", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.HeapInuse)
			m, _ := models.CreateMetrics("HeapInuse", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.HeapObjects)
			m, _ := models.CreateMetrics("HeapObjects", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.HeapReleased)
			m, _ := models.CreateMetrics("HeapReleased", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.HeapSys)
			m, _ := models.CreateMetrics("HeapSys", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.LastGC)
			m, _ := models.CreateMetrics("LastGC", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.Lookups)
			m, _ := models.CreateMetrics("Lookups", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.Mallocs)
			m, _ := models.CreateMetrics("Mallocs", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.MCacheInuse)
			m, _ := models.CreateMetrics("MCacheInuse", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.MCacheSys)
			m, _ := models.CreateMetrics("MCacheSys", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.MSpanInuse)
			m, _ := models.CreateMetrics("MSpanInuse", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.MSpanSys)
			m, _ := models.CreateMetrics("MSpanSys", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.NextGC)
			m, _ := models.CreateMetrics("NextGC", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.NumForcedGC)
			m, _ := models.CreateMetrics("NumForcedGC", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.NumGC)
			m, _ := models.CreateMetrics("NumGC", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.OtherSys)
			m, _ := models.CreateMetrics("OtherSys", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.PauseTotalNs)
			m, _ := models.CreateMetrics("PauseTotalNs", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.StackInuse)
			m, _ := models.CreateMetrics("StackInuse", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.StackSys)
			m, _ := models.CreateMetrics("StackSys", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.Sys)
			m, _ := models.CreateMetrics("Sys", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.TotalAlloc)
			m, _ := models.CreateMetrics("TotalAlloc", models.Gauge, 0, value)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			m, _ := models.CreateMetrics("PollCount", models.Counter, a.PollCount, 0)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			m, _ := models.CreateMetrics("RandomValue", models.Gauge, 0, a.RandomValue)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			m, _ := models.CreateMetrics("TotalMemory", models.Gauge, 0, a.TotalMemory)
			return m
		},
	},
	{
		Get: func(a *Agent) *models.Metrics {
			m, _ := models.CreateMetrics("FreeMemory", models.Gauge, 0, a.FreeMemory)
			return m
		},
	},
}

func (a *Agent) buildMetrics() []models.Metrics {
	result := make([]models.Metrics, 0, len(metricsRegistry)+len(a.CPUUtilizations))

	a.mu.Lock()
	for _, md := range metricsRegistry {
		m := md.Get(a)
		result = append(result, *m)
	}

	for i, val := range a.CPUUtilizations {
		coreVal := val
		metricName := fmt.Sprintf("CPUutilization%d", i+1)
		result = append(result, model.Metrics{
			ID:    metricName,
			MType: model.Gauge,
			Value: &coreVal,
		})
	}
	a.mu.Unlock()
	return result
}

func (a *Agent) sendMetricsURL(metrics []*models.Metrics) error {
	for _, m := range metrics {
		url := a.createURLFromMetric(m)
		_, err := a.client.R().
			SetHeader("Content-Type", textPlain).
			Post(url)
		if err != nil {
			return err
		}

	}

	return nil
}

func (a *Agent) sendMetricsJSON(metrics []models.Metrics, gzipCompressed bool) error {
	if len(metrics) == 0 {
		return nil
	}

	rawJSON, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	payload := rawJSON

	req := a.client.R().
		SetHeader("Content-Type", applicationJSON)

	if gzipCompressed {
		compressedBytes, err := gzipCompressJSON(rawJSON)
		if err != nil {
			return err
		}

		req.SetHeader("Content-Encoding", gzipEncoding)
		payload = compressedBytes
	}

	// check if we need hashing
	if a.cfg.Key != "" {
		key := []byte(a.cfg.Key)
		sign := hashing.Calculate(payload, key)
		encodedSign := hex.EncodeToString(sign)
		req.SetHeader("HashSHA256", encodedSign)
	}

	url := fmt.Sprintf("http://%s/updates", a.cfg.ServerAddress)
	resp, err := req.SetBody(payload).Post(url)

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("server returned status %s", resp.Status())
	}

	return nil
}

func gzipCompressJSON(body []byte) ([]byte, error) {
	var b bytes.Buffer

	gzWriter, err := gzip.NewWriterLevel(&b, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}

	_, err = gzWriter.Write(body)
	if err != nil {
		gzWriter.Close()
		return nil, err
	}

	if err := gzWriter.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (a *Agent) createURLFromMetric(m *models.Metrics) string {
	var url string
	switch m.MType {
	case models.Counter:
		url = fmt.Sprintf("http://%s/update/%s/%s/%d", a.cfg.ServerAddress, m.MType, m.ID, *m.Delta)
	case models.Gauge:
		url = fmt.Sprintf("http://%s/update/%s/%s/%.2f", a.cfg.ServerAddress, m.MType, m.ID, *m.Value)
	default:
	}

	return url
}

func calcNewRandomValue() float64 {
	rawVal := rand.Float64() * float64(rand.Int31n(100))
	return math.Round(rawVal*100) / 100
}
