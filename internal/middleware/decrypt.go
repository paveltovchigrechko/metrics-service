package middleware

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"

	"github.com/paveltovchigrechko/metrics-service/internal/encryption"
	"github.com/paveltovchigrechko/metrics-service/internal/handler"
)

// DecryptMiddleware decrypts request bodies that the agent encrypted with the
// public key matching key. Only requests with the encryption.Header header are
// decrypted.
//
// The middleware must run before GZIPMiddleware, because the agent shall compress the body
// before encrypting it, and after VerifyHashMiddleware, because the agent shall sign
// the encrypted body.
func DecryptMiddleware(key *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scheme := r.Header.Get(encryption.Header)
			if scheme == "" {
				next.ServeHTTP(w, r)
				return
			}

			if scheme != encryption.Scheme {
				handler.WriteError(w, fmt.Errorf("unsupported encryption scheme %q", scheme), http.StatusBadRequest)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				handler.WriteError(w, err, http.StatusBadRequest)
				return
			}

			plaintext, err := encryption.Decrypt(key, body)
			if err != nil {
				handler.WriteError(w, err, http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(plaintext))
			r.ContentLength = int64(len(plaintext))
			r.Header.Del(encryption.Header)

			next.ServeHTTP(w, r)
		})
	}
}
