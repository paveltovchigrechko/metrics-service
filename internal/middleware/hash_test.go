package middleware

import (
	"bytes"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paveltovchigrechko/metrics-service/internal/hashing"
)

func TestVerifyHashMiddleware(t *testing.T) {
	secret := "test-secret-key"
	keyBytes := []byte(secret)
	bodyBytes := []byte(`{"id":"PollCount","type":"counter","delta":42}`)
	validHash := hex.EncodeToString(hashing.Calculate(bodyBytes, keyBytes))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	t.Run("Empty secret key bypasses verification", func(t *testing.T) {
		mw := VerifyHashMiddleware("")(nextHandler)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()

		mw.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("Invalid hex format in HashSHA256 header returns 400", func(t *testing.T) {
		mw := VerifyHashMiddleware(secret)(nextHandler)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
		req.Header.Set("HashSHA256", "not-valid-hex-string!!!")
		rec := httptest.NewRecorder()

		mw.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 for invalid hex, got %d", rec.Code)
		}
	})

	t.Run("Hash mismatch returns 400", func(t *testing.T) {
		mw := VerifyHashMiddleware(secret)(nextHandler)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
		wrongHash := hex.EncodeToString([]byte("incorrect-hash-signature"))
		req.Header.Set("HashSHA256", wrongHash)
		rec := httptest.NewRecorder()

		mw.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 for hash mismatch, got %d", rec.Code)
		}
	})

	t.Run("Valid hash passes through successfully", func(t *testing.T) {
		mw := VerifyHashMiddleware(secret)(nextHandler)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
		req.Header.Set("HashSHA256", validHash)
		rec := httptest.NewRecorder()

		mw.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200 for valid hash, got %d", rec.Code)
		}
	})
}

func TestSignResponseMiddleware(t *testing.T) {
	secret := "test-secret-key"
	keyBytes := []byte(secret)

	t.Run("Empty secret key skips response signing", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("response body"))
		})

		mw := SignResponseMiddleware("")(nextHandler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		mw.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		if rec.Header().Get("HashSHA256") != "" {
			t.Errorf("expected no HashSHA256 header when secret is empty")
		}
	})

	t.Run("Signs response body and preserves custom status code (e.g., 404)", func(t *testing.T) {
		responseBody := []byte(`{"error":"metrics not found"}`)
		statusCode := http.StatusNotFound

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
			w.Write(responseBody)
		})

		mw := SignResponseMiddleware(secret)(nextHandler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		mw.ServeHTTP(rec, req)

		// Verify status code is preserved correctly
		if rec.Code != statusCode {
			t.Errorf("expected status %d, got %d", statusCode, rec.Code)
		}

		// Verify hash header is correctly calculated and attached
		expectedHash := hex.EncodeToString(hashing.Calculate(responseBody, keyBytes))
		actualHash := rec.Header().Get("HashSHA256")
		if actualHash != expectedHash {
			t.Errorf("expected hash %s, got %s", expectedHash, actualHash)
		}
	})

	t.Run("Implicit 200 OK status when WriteHeader is not explicitly called", func(t *testing.T) {
		responseBody := []byte("implicit status body")

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(responseBody)
		})

		mw := SignResponseMiddleware(secret)(nextHandler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		mw.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		expectedHash := hex.EncodeToString(hashing.Calculate(responseBody, keyBytes))
		actualHash := rec.Header().Get("HashSHA256")
		if actualHash != expectedHash {
			t.Errorf("expected hash %s, got %s", expectedHash, actualHash)
		}
	})
}
