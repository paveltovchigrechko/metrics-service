package agent

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	models "github.com/paveltovchigrechko/metrics-service/internal/model"
)

const (
	contentType  = "text/plain"
	host         = "http://localhost:8080"
	pollInterval = 2 * time.Second
)

type Agent struct {
	m      *runtime.MemStats
	client *http.Client

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
}

func NewAgent() *Agent {
	m := new(runtime.MemStats)
	return &Agent{
		m: m,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
		PollCount:   0,
		RandomValue: calcNewRandomValue(),
	}
}

func (a *Agent) Run() {
	for {
		runtime.ReadMemStats(a.m)
		a.updateMetrics()

		// Check if 10 seconds elapsed
		if a.PollCount%5 == 0 {
			metrics := a.buildMetrics()
			err := a.SendMetrics(metrics)
			if err != nil {
				log.Print(err)
			}
		}
		time.Sleep(pollInterval)
	}
}

func (a *Agent) updateMetrics() {
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

type MetricDescriptor struct {
	Name string
	Get  func(*Agent) *models.Metrics
}

var metricsRegistry = []MetricDescriptor{
	{
		Name: "Alloc",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.Alloc)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "BuckHashSys",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.BuckHashSys)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "Frees",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.Frees)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "GCCPUFraction",
		Get: func(a *Agent) *models.Metrics {
			value := a.GCCPUFraction
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "GCSys",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.GCSys)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "HeapAlloc",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.HeapAlloc)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "HeapIdle",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.HeapIdle)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "HeapInuse",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.HeapInuse)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "HeapObjects",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.HeapObjects)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "HeapReleased",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.HeapReleased)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "HeapSys",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.HeapSys)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "LastGC",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.LastGC)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "Lookups",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.Lookups)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "Mallocs",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.Mallocs)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "MCacheInuse",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.MCacheInuse)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "MCacheSys",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.MCacheSys)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "MSpanInuse",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.MSpanInuse)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "MSpanSys",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.MSpanSys)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "NextGC",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.NextGC)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "NumForcedGC",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.NumForcedGC)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "NumGC",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.NumGC)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "OtherSys",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.OtherSys)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "PauseTotalNs",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.PauseTotalNs)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "StackInuse",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.StackInuse)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "StackSys",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.StackSys)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "Sys",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.Sys)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "TotalAlloc",
		Get: func(a *Agent) *models.Metrics {
			value := float64(a.TotalAlloc)
			m, _ := models.CreateMetrics(models.Gauge, 0, value)
			return m
		},
	},
	{
		Name: "PollCount",
		Get: func(a *Agent) *models.Metrics {
			m, _ := models.CreateMetrics(models.Counter, a.PollCount, 0)
			return m
		},
	},
	{
		Name: "RandomValue",
		Get: func(a *Agent) *models.Metrics {
			m, _ := models.CreateMetrics(models.Gauge, 0, a.RandomValue)
			return m
		},
	},
}

func (a *Agent) buildMetrics() []*models.Metrics {
	result := make([]*models.Metrics, 0, len(metricsRegistry))

	for _, md := range metricsRegistry {
		m := md.Get(a)
		m.ID = md.Name
		result = append(result, m)
	}
	return result
}

func (a *Agent) SendMetrics(metrics []*models.Metrics) error {
	for _, m := range metrics {
		url := createURLFromMetric(m)
		resp, err := a.client.Post(url, contentType, nil)
		if err != nil {
			return err
		}

		resp.Body.Close()
	}

	return nil
}

func createURLFromMetric(m *models.Metrics) string {
	var url string
	switch m.MType {
	case models.Counter:
		url = fmt.Sprintf("%s/update/%s/%s/%d", host, m.MType, m.ID, *m.Delta)
	case models.Gauge:
		url = fmt.Sprintf("%s/update/%s/%s/%.2f", host, m.MType, m.ID, *m.Value)
	default:
	}

	return url
}

func calcNewRandomValue() float64 {
	rawVal := rand.Float64() * float64(rand.Int31n(100))
	return math.Round(rawVal*100) / 100
}
