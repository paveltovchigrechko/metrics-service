package middleware

import (
	"bytes"
	"crypto/hmac"
	"encoding/hex"
	"errors"
	"io"
	"net/http"

	"github.com/paveltovchigrechko/metrics-service/internal/handler"
	"github.com/paveltovchigrechko/metrics-service/internal/hashing"
)

var (
	ErrMissingHash  = errors.New("HashSHA256 header contains no hash")
	ErrHashMismatch = errors.New("header hash does not match body hash")
)

type hmacWriter struct {
	http.ResponseWriter
	body          *bytes.Buffer
	statusCode    int
	headerWritten bool
}

func (hw *hmacWriter) Write(b []byte) (int, error) {
	if !hw.headerWritten {
		hw.WriteHeader(http.StatusOK) // Implicitly set 200 OK if Write is called first
	}
	return hw.body.Write(b)
}

func (hw *hmacWriter) WriteHeader(statusCode int) {
	if hw.headerWritten {
		return // net/http ignores subsequent WriteHeader calls
	}
	hw.statusCode = statusCode
	hw.headerWritten = true
}

// VerifyHashMiddleware verifies the incoming request body's HMAC signature.
func VerifyHashMiddleware(secretKey string) func(http.Handler) http.Handler {
	if secretKey == "" {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	keyBytes := []byte(secretKey)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hashHeader := r.Header.Get("HashSHA256")
			if hashHeader == "" {
				// Coupling with handler package essentially is not the best idea.
				handler.WriteError(w, ErrMissingHash, http.StatusBadRequest)
				return
			}

			receivedHash, err := hex.DecodeString(hashHeader)
			if err != nil {
				handler.WriteError(w, err, http.StatusBadRequest)
				return
			}

			body, err := io.ReadAll(r.Body) // Consumes the body
			if err != nil {
				handler.WriteError(w, err, http.StatusBadRequest)
				return
			}
			// Restore the body
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			expectedHash := hashing.Calculate(body, keyBytes)
			if !hmac.Equal(receivedHash, expectedHash) {
				handler.WriteError(w, ErrHashMismatch, http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SignResponseMiddleware computes and attaches the HMAC signature to the response header.
func SignResponseMiddleware(secretKey string) func(http.Handler) http.Handler {
	if secretKey == "" {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	keyBytes := []byte(secretKey)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hw := &hmacWriter{
				ResponseWriter: w,
				body:           bytes.NewBuffer(nil),
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(hw, r)

			responseBody := hw.body.Bytes()
			responseHash := hashing.Calculate(responseBody, keyBytes)

			w.Header().Set("HashSHA256", hex.EncodeToString(responseHash))

			code := hw.statusCode
			if !hw.headerWritten {
				code = http.StatusOK
			}

			w.WriteHeader(code)
			w.Write(responseBody)
		})
	}
}
