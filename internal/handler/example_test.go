package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/paveltovchigrechko/metrics-service/internal/handler"
	"github.com/paveltovchigrechko/metrics-service/internal/model"
	"github.com/paveltovchigrechko/metrics-service/internal/repository"
)

// Update a single metric with its type, name and value in the URL path.
func ExampleAppHandler_PostMetrics() {
	storage := repository.NewMemStorage()
	h := handler.NewHandler(storage, nil, nil, nil)

	r := chi.NewRouter()
	r.Post("/update/{metricsType}/{metricsName}/{metricsValue}", h.PostMetrics)

	req := httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/5", nil)
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	fmt.Println(rec.Code)

	m, _ := storage.GetMetrics(context.Background(), "PollCount", model.Counter)
	fmt.Println(m.ID, *m.Delta)

	// Output:
	// 200
	// PollCount 5
}

// Read the value of a single metric as plain text.
func ExampleAppHandler_MetricsValue() {
	storage := repository.NewMemStorage()
	alloc, _ := model.CreateMetrics("Alloc", model.Gauge, 0, 123.45)
	_ = storage.SaveMetrics(context.Background(), alloc)
	h := handler.NewHandler(storage, nil, nil, nil)

	r := chi.NewRouter()
	r.Get("/value/{metricsType}/{metricsName}", h.MetricsValue)

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/Alloc", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())

	// Output:
	// 200
	// 123.45
}

// Update a single metric sent as JSON. A counter is set with "delta",
// a gauge with "value".
func ExampleAppHandler_UpdateEndpoint() {
	storage := repository.NewMemStorage()
	h := handler.NewHandler(storage, nil, nil, nil)

	r := chi.NewRouter()
	r.Post("/update", h.UpdateEndpoint)

	body := `{"id":"Alloc","type":"gauge","value":123.45}`
	req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	fmt.Println(rec.Code)

	m, _ := storage.GetMetrics(context.Background(), "Alloc", model.Gauge)
	fmt.Println(m.ID, *m.Value)

	// Output:
	// 200
	// Alloc 123.45
}

// Update several metrics at once with a JSON array.
func ExampleAppHandler_UpdatesEndpoint() {
	storage := repository.NewMemStorage()
	h := handler.NewHandler(storage, nil, nil, nil)

	r := chi.NewRouter()
	r.Post("/updates", h.UpdatesEndpoint)

	body := `[
		{"id":"PollCount","type":"counter","delta":5},
		{"id":"Alloc","type":"gauge","value":123.45}
	]`
	req := httptest.NewRequest(http.MethodPost, "/updates", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	fmt.Println(rec.Code)

	counter, _ := storage.GetMetrics(context.Background(), "PollCount", model.Counter)
	gauge, _ := storage.GetMetrics(context.Background(), "Alloc", model.Gauge)
	fmt.Println(counter.ID, *counter.Delta)
	fmt.Println(gauge.ID, *gauge.Value)

	// Output:
	// 200
	// PollCount 5
	// Alloc 123.45
}

// Request a metric by its name and type as JSON. The response contains the
// metric with its current value.
func ExampleAppHandler_ValueEndpoint() {
	storage := repository.NewMemStorage()
	pollCount, _ := model.CreateMetrics("PollCount", model.Counter, 5, 0)
	_ = storage.SaveMetrics(context.Background(), pollCount)
	h := handler.NewHandler(storage, nil, nil, nil)

	r := chi.NewRouter()
	r.Post("/value", h.ValueEndpoint)

	body := `{"id":"PollCount","type":"counter"}`
	req := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("Content-Type"))
	fmt.Println(rec.Body.String())

	// Output:
	// 200
	// application/json
	// {"id":"PollCount","type":"counter","delta":5}
}

// Render all stored metrics as an HTML page.
func ExampleAppHandler_MainPage() {
	storage := repository.NewMemStorage()
	pollCount, _ := model.CreateMetrics("PollCount", model.Counter, 5, 0)
	_ = storage.SaveMetrics(context.Background(), pollCount)
	h := handler.NewHandler(storage, nil, nil, nil)

	r := chi.NewRouter()
	r.Get("/", h.MainPage)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("Content-Type"))
	fmt.Println(strings.Contains(rec.Body.String(), "PollCount&nbsp;&nbsp;&nbsp;&nbsp;5"))

	// Output:
	// 200
	// text/html; charset=utf-8
	// true
}

type databasePinger struct{}

func (databasePinger) PingContext(context.Context) error { return nil }

// Check that the database is reachable. Without a configured database the
// endpoint responds with 500 Internal Server Error.
func ExampleAppHandler_PingEndpoint() {
	h := handler.NewHandler(repository.NewMemStorage(), databasePinger{}, nil, nil)

	r := chi.NewRouter()
	r.Get("/ping", h.PingEndpoint)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	fmt.Println(rec.Code)

	// Output:
	// 200
}
