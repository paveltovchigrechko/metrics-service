package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paveltovchigrechko/metrics-service/internal/config"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to construct a pointer value
func int64Ptr(v int64) *int64 { return &v }

func TestNewServer(t *testing.T) {
	t.Run("Standard initialization without restore", func(t *testing.T) {
		cfg := &config.ServerConfig{
			ServerAddress:   "localhost:8080",
			StoreInterval:   time.Second * 5,
			FileStoragePath: filepath.Join(t.TempDir(), "metrics.json"),
			Restore:         false,
		}

		srv, err := New(cfg)
		require.NoError(t, err)
		assert.NotNil(t, srv)
		assert.NotNil(t, srv.storage)
		assert.NotNil(t, srv.handler)
		assert.NotNil(t, srv.router)
	})

	t.Run("Initialization with Restore from existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		backupFile := filepath.Join(tmpDir, "metrics.json")

		historicalMetrics := []models.Metrics{
			{
				ID:    "ExistingCounter",
				MType: models.Counter,
				Delta: int64Ptr(120),
			},
		}
		bytes, err := json.Marshal(historicalMetrics)
		require.NoError(t, err)
		err = os.WriteFile(backupFile, bytes, 0644)
		require.NoError(t, err)

		cfg := &config.ServerConfig{
			ServerAddress:   "localhost:8080",
			StoreInterval:   time.Second * 5,
			FileStoragePath: backupFile,
			Restore:         true,
		}

		srv, err := New(cfg)
		require.NoError(t, err)
		assert.NotNil(t, srv)

		restoredMetric, getErr := srv.storage.GetMetrics(context.Background(), "ExistingCounter", models.Counter)
		require.NoError(t, getErr)
		require.NotNil(t, restoredMetric)
		assert.Equal(t, int64(120), *restoredMetric.Delta)
	})

	t.Run("Accepts non-existent file gracefully on first run with Restore true", func(t *testing.T) {
		cfg := &config.ServerConfig{
			ServerAddress:   "localhost:8080",
			StoreInterval:   time.Second * 5,
			FileStoragePath: filepath.Join(t.TempDir(), "this-file-does-not-exist.json"),
			Restore:         true,
		}

		srv, err := New(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, srv)
	})

	t.Run("Initialization fails on a corrupted restore file format", func(t *testing.T) {
		tmpDir := t.TempDir()
		corruptFile := filepath.Join(tmpDir, "metrics.json")
		err := os.WriteFile(corruptFile, []byte(`{invalid-json`), 0644)
		require.NoError(t, err)

		cfg := &config.ServerConfig{
			ServerAddress:   "localhost:8080",
			StoreInterval:   time.Second * 5,
			FileStoragePath: corruptFile,
			Restore:         true,
		}

		srv, err := New(cfg)
		assert.Error(t, err)
		assert.Nil(t, srv)
	})

	t.Run("Synchronous saving strategy setup (StoreInterval = 0)", func(t *testing.T) {
		cfg := &config.ServerConfig{
			ServerAddress:   "localhost:8080",
			StoreInterval:   0, // Synchronous
			FileStoragePath: filepath.Join(t.TempDir(), "sync-metrics.json"),
			Restore:         false,
		}

		srv, err := New(cfg)
		require.NoError(t, err)
		assert.NotNil(t, srv)

		req := httptest.NewRequest(http.MethodPost, "/update/counter/Hits/5", nil)
		w := httptest.NewRecorder()
		srv.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify that it synchronously wrote directly to the filesystem
		bytes, readErr := os.ReadFile(cfg.FileStoragePath)
		assert.NoError(t, readErr)

		var saved []models.Metrics
		assert.NoError(t, json.Unmarshal(bytes, &saved))
		assert.Len(t, saved, 1)
		assert.Equal(t, "Hits", saved[0].ID)
		assert.Equal(t, int64(5), *saved[0].Delta)
	})
}

func TestServer_useMiddlewares(t *testing.T) {
	cfg := &config.ServerConfig{
		ServerAddress:   "localhost:8080",
		StoreInterval:   time.Second * 5,
		FileStoragePath: filepath.Join(t.TempDir(), "metrics.json"),
		Restore:         false,
	}

	middlewareTriggered := false
	dummyMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareTriggered = true
			next.ServeHTTP(w, r)
		})
	}

	srv, err := New(cfg, dummyMiddleware)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	assert.True(t, middlewareTriggered, "middleware chain should be reached during processing")
}

func TestServer_setHandlers(t *testing.T) {
	cfg := &config.ServerConfig{
		ServerAddress:   "localhost:8080",
		StoreInterval:   time.Second * 5,
		FileStoragePath: filepath.Join(t.TempDir(), "metrics.json"),
		Restore:         false,
	}

	srv, err := New(cfg)
	require.NoError(t, err)

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/update/gauge/Alloc/12.34"},
		{http.MethodPost, "/update"},
		{http.MethodPost, "/update/"},
		{http.MethodPost, "/value"},
		{http.MethodPost, "/value/"},
		{http.MethodGet, "/"},
		{http.MethodGet, "/value/gauge/Alloc"},
	}

	// Verify all configured endpoints are registered inside the router multiplexer
	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()
			srv.router.ServeHTTP(w, req)

			// We don't necessarily care about 200 OK because values may be missing,
			// we just want to verify it didn't trigger a 404 (Not Found) or 405 (Method Not Allowed).
			assert.NotEqual(t, http.StatusNotFound, w.Code)
		})
	}
}

func TestServer_runStoreLoop(t *testing.T) {
	tmpDir := t.TempDir()
	backupPath := filepath.Join(tmpDir, "loop-metrics.json")

	cfg := &config.ServerConfig{
		ServerAddress:   "localhost:8080",
		StoreInterval:   time.Millisecond * 10, // Fast interval for testing
		FileStoragePath: backupPath,
		Restore:         false,
	}

	srv, err := New(cfg)
	require.NoError(t, err)

	// Inject a metric into local memory store
	err = srv.storage.SaveMetrics(context.Background(), &models.Metrics{
		ID:    "ActiveSessions",
		MType: models.Counter,
		Delta: int64Ptr(777),
	})
	require.NoError(t, err)

	// Start the storage loop in a separate goroutine
	go srv.runStoreLoop()

	// Poll the file until it has been successfully written and contains the data we expect
	assert.Eventually(t, func() bool {
		bytes, err := os.ReadFile(backupPath)
		if err != nil {
			return false // File doesn't exist yet or is currently locked
		}

		var saved []models.Metrics
		if err := json.Unmarshal(bytes, &saved); err != nil {
			return false // JSON is empty/corrupt or still being written
		}

		// Ensure we parsed the metrics correctly and the specific one is present
		if len(saved) == 1 && saved[0].ID == "ActiveSessions" && *saved[0].Delta == 777 {
			return true
		}

		return false
	}, time.Second*2, time.Millisecond*5, "The storage loop did not save the metrics correctly to disk in time")
}
