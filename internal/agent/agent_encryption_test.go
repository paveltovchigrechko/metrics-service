package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paveltovchigrechko/metrics-service/internal/config"
	"github.com/paveltovchigrechko/metrics-service/internal/encryption"
	"github.com/paveltovchigrechko/metrics-service/internal/encryption/encryptiontest"
	"github.com/paveltovchigrechko/metrics-service/internal/hashing"
	"github.com/paveltovchigrechko/metrics-service/internal/middleware"
	"github.com/paveltovchigrechko/metrics-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type capturedRequest struct {
	header http.Header
	body   []byte
}

// newCaptureServer starts a server that records the last request it received.
func newCaptureServer(t *testing.T) (addr string, last func() capturedRequest) {
	t.Helper()

	var got capturedRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		got = capturedRequest{header: r.Header.Clone(), body: body}
	}))
	t.Cleanup(server.Close)

	return strings.TrimPrefix(server.URL, "http://"), func() capturedRequest { return got }
}

func newTestAgent(t *testing.T, args ...string) *Agent {
	t.Helper()

	cfg, err := config.SetAgentConfig(args)
	require.NoError(t, err)
	a, err := NewAgent(cfg)
	require.NoError(t, err)

	return a
}

func gunzip(t *testing.T, data []byte) []byte {
	t.Helper()

	zr, err := gzip.NewReader(bytes.NewReader(data))
	require.NoError(t, err)
	out, err := io.ReadAll(zr)
	require.NoError(t, err)

	return out
}

func testMetrics(t *testing.T) []model.Metrics {
	t.Helper()

	counter, err := model.CreateMetrics("PollCount", model.Counter, 3, 0)
	require.NoError(t, err)
	gauge, err := model.CreateMetrics("Alloc", model.Gauge, 0, 42.5)
	require.NoError(t, err)

	return []model.Metrics{*counter, *gauge}
}

func TestNewAgent_CryptoKey(t *testing.T) {
	t.Run("loads the public key", func(t *testing.T) {
		publicPath, _, key := encryptiontest.WriteKeyFiles(t)

		a := newTestAgent(t, "-crypto-key", publicPath)

		require.NotNil(t, a.publicKey)
		assert.True(t, a.publicKey.Equal(&key.PublicKey))
	})

	t.Run("encryption is disabled without a key", func(t *testing.T) {
		a := newTestAgent(t)

		assert.Nil(t, a.publicKey)
	})

	t.Run("fails if the key file is missing", func(t *testing.T) {
		cfg, err := config.SetAgentConfig([]string{"-crypto-key", filepath.Join(t.TempDir(), "missing.pem")})
		require.NoError(t, err)

		a, err := NewAgent(cfg)

		assert.Error(t, err)
		assert.Nil(t, a)
	})

	t.Run("fails if the file holds a private key", func(t *testing.T) {
		_, privatePath, _ := encryptiontest.WriteKeyFiles(t)
		cfg, err := config.SetAgentConfig([]string{"-crypto-key", privatePath})
		require.NoError(t, err)

		_, err = NewAgent(cfg)

		assert.Error(t, err)
	})
}

func TestSendMetricsJSON_Encryption(t *testing.T) {
	metrics := testMetrics(t)
	wantJSON, err := json.Marshal(metrics)
	require.NoError(t, err)

	t.Run("encrypts the gzipped body with the public key", func(t *testing.T) {
		addr, last := newCaptureServer(t)
		publicPath, _, key := encryptiontest.WriteKeyFiles(t)
		a := newTestAgent(t, "-a", addr, "-crypto-key", publicPath)

		require.NoError(t, a.sendMetricsJSON(metrics, true))

		got := last()
		assert.Equal(t, encryption.Scheme, got.header.Get(encryption.Header))
		assert.Equal(t, "gzip", got.header.Get("Content-Encoding"))
		assert.NotContains(t, string(got.body), "PollCount")

		plaintext, err := encryption.Decrypt(key, got.body)
		require.NoError(t, err)
		assert.JSONEq(t, string(wantJSON), string(gunzip(t, plaintext)))
	})

	t.Run("encrypts an uncompressed body", func(t *testing.T) {
		addr, last := newCaptureServer(t)
		publicPath, _, key := encryptiontest.WriteKeyFiles(t)
		a := newTestAgent(t, "-a", addr, "-crypto-key", publicPath)

		require.NoError(t, a.sendMetricsJSON(metrics, false))

		plaintext, err := encryption.Decrypt(key, last().body)
		require.NoError(t, err)
		assert.JSONEq(t, string(wantJSON), string(plaintext))
	})

	t.Run("signs the encrypted body", func(t *testing.T) {
		const secret = "secret"
		addr, last := newCaptureServer(t)
		publicPath, _, _ := encryptiontest.WriteKeyFiles(t)
		a := newTestAgent(t, "-a", addr, "-k", secret, "-crypto-key", publicPath)

		require.NoError(t, a.sendMetricsJSON(metrics, true))

		got := last()
		sign, err := base64.StdEncoding.DecodeString(got.header.Get("HashSHA256"))
		require.NoError(t, err)
		assert.True(t, hmac.Equal(hashing.Calculate(got.body, []byte(secret)), sign))
	})

	t.Run("server middleware chain accepts the request", func(t *testing.T) {
		const secret = "secret"
		publicPath, _, key := encryptiontest.WriteKeyFiles(t)

		var received []model.Metrics
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		})
		chain := middleware.VerifyHashMiddleware(secret)(
			middleware.DecryptMiddleware(key)(
				middleware.GZIPMiddleware(handler)))
		server := httptest.NewServer(chain)
		t.Cleanup(server.Close)

		a := newTestAgent(t, "-a", strings.TrimPrefix(server.URL, "http://"), "-k", secret, "-crypto-key", publicPath)

		require.NoError(t, a.sendMetricsJSON(metrics, true))
		assert.Equal(t, metrics, received)
	})

	t.Run("server with another private key rejects the request", func(t *testing.T) {
		publicPath, _, _ := encryptiontest.WriteKeyFiles(t)
		otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		server := httptest.NewServer(middleware.DecryptMiddleware(otherKey)(http.NotFoundHandler()))
		t.Cleanup(server.Close)

		a := newTestAgent(t, "-a", strings.TrimPrefix(server.URL, "http://"), "-crypto-key", publicPath)

		assert.Error(t, a.sendMetricsJSON(metrics, true))
	})

	t.Run("sends plain gzip without a key", func(t *testing.T) {
		addr, last := newCaptureServer(t)
		a := newTestAgent(t, "-a", addr)

		require.NoError(t, a.sendMetricsJSON(metrics, true))

		got := last()
		assert.Empty(t, got.header.Get(encryption.Header))
		assert.JSONEq(t, string(wantJSON), string(gunzip(t, got.body)))
	})
}
