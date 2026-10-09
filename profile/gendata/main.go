// Command gendata generates deterministic test data for load testing and profiling the metrics server.
//
// Usage (from the repository root):
//
//	go run ./profile/gendata -out profile
//
// Generated files:
//
//	update_gauge.json      single gauge for POST /update
//	update_counter.json    single counter for POST /update
//	updates_small.json     batch like the agent sends (all runtime gauges + PollCount + RandomValue) for POST /updates
//	updates_large.json     big batch (-large metrics) for POST /updates
//	value_gauge.json       lookup request for POST /value
//	value_counter.json     lookup request for POST /value
//	seed_metrics.json      server state (-seed metrics) to restore on start: server -f <copy of file> -r
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/paveltovchigrechko/metrics-service/internal/model"
)

// Names of the gauges the agent reports from runtime.MemStats.
var runtimeGauges = []string{
	"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse",
	"HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
	"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys",
	"PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue",
}

const pollCount = "PollCount"

func main() {
	out := flag.String("out", "profile", "output directory")
	seedSize := flag.Int("seed", 5000, "number of metrics in seed_metrics.json")
	largeSize := flag.Int("large", 1000, "number of metrics in updates_large.json")
	flag.Parse()

	rnd := rand.New(rand.NewSource(1)) // Fixed seed: same data on every run.

	files := map[string]any{
		"update_gauge.json":   gauge("Alloc", rnd),
		"update_counter.json": counter(pollCount, rnd),
		"updates_small.json":  agentBatch(rnd),
		"updates_large.json":  syntheticMetrics(*largeSize, rnd),
		"value_gauge.json":    map[string]string{"id": "Alloc", "type": model.Gauge},
		"value_counter.json":  map[string]string{"id": pollCount, "type": model.Counter},
		"seed_metrics.json":   syntheticMetrics(*seedSize, rnd),
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatal(err)
	}

	for name, data := range files {
		if err := write(filepath.Join(*out, name), data); err != nil {
			log.Fatal(err)
		}
	}
}

func gauge(id string, rnd *rand.Rand) model.Metrics {
	v := roundValue(rnd.Float64() * 1e6)
	return model.Metrics{ID: id, MType: model.Gauge, Value: &v}
}

func counter(id string, rnd *rand.Rand) model.Metrics {
	d := int64(rnd.Intn(10) + 1)
	return model.Metrics{ID: id, MType: model.Counter, Delta: &d}
}

// agentBatch mimics a single agent report.
func agentBatch(rnd *rand.Rand) []model.Metrics {
	batch := make([]model.Metrics, 0, len(runtimeGauges)+1)
	for _, name := range runtimeGauges {
		batch = append(batch, gauge(name, rnd))
	}

	return append(batch, counter(pollCount, rnd))
}

// syntheticMetrics returns n unique metrics (about 3 gauges per counter): gauge_00000, counter_00001, ...
func syntheticMetrics(n int, rnd *rand.Rand) []model.Metrics {
	metrics := make([]model.Metrics, 0, n)
	for i := 0; i < n; i++ {
		if i%4 == 3 {
			metrics = append(metrics, counter(fmt.Sprintf("counter_%05d", i), rnd))
		} else {
			metrics = append(metrics, gauge(fmt.Sprintf("gauge_%05d", i), rnd))
		}
	}

	return metrics
}

func roundValue(v float64) float64 {
	return float64(int64(v*1000)) / 1000
}

func write(path string, data any) error {
	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(path, encoded, 0o644)
}
