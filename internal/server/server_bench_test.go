package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paveltovchigrechko/metrics-service/internal/config"
	"github.com/paveltovchigrechko/metrics-service/internal/hashing"
	"github.com/paveltovchigrechko/metrics-service/internal/logger"
	"github.com/paveltovchigrechko/metrics-service/internal/middleware"
	"github.com/paveltovchigrechko/metrics-service/internal/model"
	"go.uber.org/zap"
)

const benchKey = "bench-secret-key"

type benchStack struct {
	name        string
	middlewares func() []func(http.Handler) http.Handler
	gzip        bool // Send Accept-Encoding: gzip
	sign        bool // Send the HashSHA256 header
}

var benchStacks = []benchStack{
	{
		name:        "bare",
		middlewares: func() []func(http.Handler) http.Handler { return nil },
	},
	{
		name: "logger",
		middlewares: func() []func(http.Handler) http.Handler {
			return []func(http.Handler) http.Handler{logger.LoggerMiddleware(zap.NewNop().Sugar())}
		},
	},
	{
		name: "logger+gzip",
		middlewares: func() []func(http.Handler) http.Handler {
			return []func(http.Handler) http.Handler{
				logger.LoggerMiddleware(zap.NewNop().Sugar()),
				middleware.GZIPMiddleware,
			}
		},
		gzip: true,
	},
	{
		name: "logger+hash+gzip",
		middlewares: func() []func(http.Handler) http.Handler {
			return []func(http.Handler) http.Handler{
				logger.LoggerMiddleware(zap.NewNop().Sugar()),
				middleware.VerifyHashMiddleware(benchKey),
				middleware.SignResponseMiddleware(benchKey),
				middleware.GZIPMiddleware,
			}
		},
		gzip: true,
		sign: true,
	},
}

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

func newBenchServer(b *testing.B, stack benchStack, seed int) *Server {
	b.Helper()

	// No file, no database, no audit: in-memory storage only.
	srv, err := New(&config.ServerConfig{ServerAddress: "localhost:8080"}, stack.middlewares()...)
	if err != nil {
		b.Fatal(err)
	}
	if err = srv.storage.SaveBatch(context.Background(), benchMetrics(seed)); err != nil {
		b.Fatal(err)
	}

	return srv
}

type discardWriter struct {
	header http.Header
	status int
	bytes  int
}

func (w *discardWriter) Header() http.Header { return w.header }

func (w *discardWriter) WriteHeader(status int) { w.status = status }

func (w *discardWriter) Write(p []byte) (int, error) {
	w.bytes += len(p)
	return len(p), nil
}

func (w *discardWriter) reset() {
	clear(w.header)
	w.status, w.bytes = 0, 0
}

func BenchmarkRouterMainPage(b *testing.B) {
	for _, stack := range benchStacks {
		b.Run(stack.name, func(b *testing.B) {
			srv := newBenchServer(b, stack, 5000)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if stack.gzip {
				req.Header.Set("Accept-Encoding", "gzip")
			}
			w := &discardWriter{header: make(http.Header)}

			b.ReportAllocs()
			for b.Loop() {
				w.reset()
				srv.router.ServeHTTP(w, req)
			}

			if w.status != 0 && w.status != http.StatusOK {
				b.Fatalf("unexpected status %d", w.status)
			}
			b.ReportMetric(float64(w.bytes), "respbytes/op")
		})
	}
}

func BenchmarkRouterUpdates(b *testing.B) {
	for _, n := range []int{29, 1000} {
		payload, err := json.Marshal(benchMetrics(n))
		if err != nil {
			b.Fatal(err)
		}
		signature := base64.StdEncoding.EncodeToString(hashing.Calculate(payload, []byte(benchKey)))

		for _, stack := range benchStacks {
			b.Run(fmt.Sprintf("metrics=%d/%s", n, stack.name), func(b *testing.B) {
				srv := newBenchServer(b, stack, 0)
				req := httptest.NewRequest(http.MethodPost, "/updates", nil)
				req.Header.Set("Content-Type", "application/json")
				if stack.gzip {
					req.Header.Set("Accept-Encoding", "gzip")
				}
				if stack.sign {
					req.Header.Set("HashSHA256", signature)
				}
				w := &discardWriter{header: make(http.Header)}

				b.SetBytes(int64(len(payload)))
				b.ReportAllocs()
				for b.Loop() {
					w.reset()
					req.Body = io.NopCloser(bytes.NewReader(payload))
					srv.router.ServeHTTP(w, req)
				}

				if w.status != 0 && w.status != http.StatusOK {
					b.Fatalf("unexpected status %d", w.status)
				}
			})
		}
	}
}
