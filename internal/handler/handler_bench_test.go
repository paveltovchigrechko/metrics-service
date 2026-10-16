package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/paveltovchigrechko/metrics-service/internal/model"
	"github.com/paveltovchigrechko/metrics-service/internal/repository"
)

func benchMetrics(n int) []model.Metrics {
	metrics := make([]model.Metrics, 0, n)
	for i := 0; i < n; i++ {
		if i%4 == 3 {
			d := int64(i)
			metrics = append(metrics, model.Metrics{ID: fmt.Sprintf("counter_%05d", i), MType: model.Counter, Delta: &d})
		} else {
			v := float64(i) * 1.5
			metrics = append(metrics, model.Metrics{ID: fmt.Sprintf("gauge_%05d", i), MType: model.Gauge, Value: &v})
		}
	}

	return metrics
}

func newBenchHandler(b *testing.B, n int) *AppHandler {
	b.Helper()

	s := repository.NewMemStorage()
	if err := s.SaveBatch(context.Background(), benchMetrics(n)); err != nil {
		b.Fatal(err)
	}

	return NewHandler(s, nil, nil, nil)
}

type countingWriter struct {
	header http.Header
	status int
	writes int
	bytes  int
}

func newCountingWriter() *countingWriter {
	return &countingWriter{header: make(http.Header)}
}

func (w *countingWriter) Header() http.Header { return w.header }

func (w *countingWriter) WriteHeader(status int) { w.status = status }

func (w *countingWriter) Write(p []byte) (int, error) {
	w.writes++
	w.bytes += len(p)

	return len(p), nil
}

func (w *countingWriter) reset() {
	clear(w.header)
	w.status, w.writes, w.bytes = 0, 0, 0
}

func jsonRequest(target string, payload []byte) (*http.Request, func()) {
	req := httptest.NewRequest(http.MethodPost, target, nil)
	req.Header.Set(contentType, applicationJSON)
	replay := func() { req.Body = io.NopCloser(bytes.NewReader(payload)) }
	replay()

	return req, replay
}

var mainPageSizes = []int{100, 1000, 5000}

func BenchmarkMainPage(b *testing.B) {
	for _, n := range mainPageSizes {
		b.Run(fmt.Sprintf("metrics=%d", n), func(b *testing.B) {
			h := newBenchHandler(b, n)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := newCountingWriter()

			b.ReportAllocs()
			for b.Loop() {
				w.reset()
				h.MainPage(w, req)
			}

			if w.status != 0 && w.status != http.StatusOK {
				b.Fatalf("unexpected status %d", w.status)
			}
			b.ReportMetric(float64(w.writes), "writes/op")
			b.ReportMetric(float64(w.bytes), "bytes/op")
		})
	}
}

func BenchmarkMainPageHTTP(b *testing.B) {
	for _, n := range mainPageSizes {
		b.Run(fmt.Sprintf("metrics=%d", n), func(b *testing.B) {
			h := newBenchHandler(b, n)
			srv := httptest.NewServer(http.HandlerFunc(h.MainPage))
			defer srv.Close()

			client := srv.Client()
			b.ReportAllocs()
			for b.Loop() {
				resp, err := client.Get(srv.URL)
				if err != nil {
					b.Fatal(err)
				}
				if _, err = io.Copy(io.Discard, resp.Body); err != nil {
					b.Fatal(err)
				}
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					b.Fatalf("unexpected status %d", resp.StatusCode)
				}
			}
		})
	}
}

func BenchmarkMainPageParallel(b *testing.B) {
	h := newBenchHandler(b, 5000)

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := newCountingWriter()
		for pb.Next() {
			w.reset()
			h.MainPage(w, req)
		}
	})
}

func BenchmarkUpdatesEndpoint(b *testing.B) {
	for _, n := range []int{29, 1000} {
		b.Run(fmt.Sprintf("metrics=%d", n), func(b *testing.B) {
			h := newBenchHandler(b, 0)
			payload, err := json.Marshal(benchMetrics(n))
			if err != nil {
				b.Fatal(err)
			}
			req, replay := jsonRequest("/updates", payload)
			w := newCountingWriter()

			b.SetBytes(int64(len(payload)))
			b.ReportAllocs()
			for b.Loop() {
				w.reset()
				replay()
				h.UpdatesEndpoint(w, req)
			}

			if w.status != 0 && w.status != http.StatusOK {
				b.Fatalf("unexpected status %d", w.status)
			}
		})
	}
}

func BenchmarkUpdateEndpoint(b *testing.B) {
	h := newBenchHandler(b, 0)
	req, replay := jsonRequest("/update", []byte(`{"id":"PollCount","type":"counter","delta":8}`))
	w := newCountingWriter()

	b.ReportAllocs()
	for b.Loop() {
		w.reset()
		replay()
		h.UpdateEndpoint(w, req)
	}

	if w.status != 0 && w.status != http.StatusOK {
		b.Fatalf("unexpected status %d", w.status)
	}
}

func BenchmarkValueEndpoint(b *testing.B) {
	h := newBenchHandler(b, 1000)
	req, replay := jsonRequest("/value", []byte(`{"id":"`+"gauge_00000"+`","type":"gauge"}`))
	w := newCountingWriter()

	b.ReportAllocs()
	for b.Loop() {
		w.reset()
		replay()
		h.ValueEndpoint(w, req)
	}

	if w.status != 0 && w.status != http.StatusOK {
		b.Fatalf("unexpected status %d (%s)", w.status, strconv.Itoa(w.bytes))
	}
}
