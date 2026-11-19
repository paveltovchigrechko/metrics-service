package middleware

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paveltovchigrechko/metrics-service/internal/encryption"
	"github.com/paveltovchigrechko/metrics-service/internal/encryption/encryptiontest"
	"github.com/paveltovchigrechko/metrics-service/internal/hashing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// echoBody responds with the request body and the encryption header it saw.
var echoBody = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("X-Seen-Encryption", r.Header.Get(encryption.Header))
	w.Write(body)
})

func TestDecryptMiddleware(t *testing.T) {
	key := encryptiontest.Key(t)
	mw := DecryptMiddleware(key)(echoBody)
	plaintext := []byte(`[{"id":"PollCount","type":"counter","delta":5}]`)

	t.Run("decrypts an encrypted body", func(t *testing.T) {
		msg, err := encryption.Encrypt(&key.PublicKey, plaintext)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(msg))
		req.Header.Set(encryption.Header, encryption.Scheme)
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, plaintext, rec.Body.Bytes())
		assert.Empty(t, rec.Header().Get("X-Seen-Encryption"), "the header is removed after decryption")
	})

	t.Run("passes requests without the header unchanged", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(plaintext))
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, plaintext, rec.Body.Bytes())
	})

	t.Run("rejects an unknown scheme", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(plaintext))
		req.Header.Set(encryption.Header, "rot13")
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("rejects a body that is not encrypted for the key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(plaintext))
		req.Header.Set(encryption.Header, encryption.Scheme)
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.NotContains(t, rec.Body.String(), string(plaintext))
	})
}

// TestDecryptMiddleware_InServerChain checks the order used in cmd/server: the
// agent gzips, encrypts and then signs the body, so the server verifies the
// hash, decrypts and then decompresses.
func TestDecryptMiddleware_InServerChain(t *testing.T) {
	key := encryptiontest.Key(t)
	const secret = "secret"
	plaintext := []byte(`[{"id":"Alloc","type":"gauge","value":1.5}]`)

	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	_, err := zw.Write(plaintext)
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	msg, err := encryption.Encrypt(&key.PublicKey, gz.Bytes())
	require.NoError(t, err)
	sign := base64.StdEncoding.EncodeToString(hashing.Calculate(msg, []byte(secret)))

	chain := VerifyHashMiddleware(secret)(DecryptMiddleware(key)(GZIPMiddleware(echoBody)))

	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(msg))
	req.Header.Set(encryption.Header, encryption.Scheme)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("HashSHA256", sign)
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, plaintext, rec.Body.Bytes())
}
