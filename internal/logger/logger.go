package logger

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
		err    error
	}

	loggingResponseWriter struct {
		http.ResponseWriter               // This http.ResponseWriter is used in our handler.
		responseData        *responseData // Additional structure to save response status and size.
	}
)

// Error interface
type ErrorRecorder interface {
	SetError(error)
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	if r.responseData.status == 0 { // Set StatusOK if we writing something to the response.
		r.WriteHeader(http.StatusOK)
	}
	size, err := r.ResponseWriter.Write(b) // We use the underline original http.ResponseWriter here!
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	if r.responseData.status != 0 { // If we have status already, discard the following.
		return
	}
	r.responseData.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode) // We use the underline original http.ResponseWriter here!
}

func (r *loggingResponseWriter) SetError(err error) {
	r.responseData.err = err
}

func New() (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	sugar := logger.Sugar()

	return sugar, nil
}

const MaxLoggedBodySize = 512

func peekBody(body io.ReadCloser, limit int) (head []byte, truncated bool, restored io.ReadCloser, err error) {
	buf := make([]byte, limit+1)
	n, err := io.ReadFull(body, buf)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		err = nil // The body is shorter than the buffer, that's fine.
	}

	restored = struct {
		io.Reader
		io.Closer
	}{io.MultiReader(bytes.NewReader(buf[:n]), body), body}

	if n > limit {
		return buf[:limit], true, restored, err
	}

	return buf[:n], false, restored, err
}

func LoggerMiddleware(log *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { // This function is basically a cast to get the required signature for func (mx *chi.Mux) Use(middlewares ...func(http.Handler) http.Handler)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			responseData := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}

			var bodyHead []byte
			var bodyTruncated bool
			if r.Body != nil && r.Body != http.NoBody {
				var err error
				bodyHead, bodyTruncated, r.Body, err = peekBody(r.Body, MaxLoggedBodySize)
				if err != nil {
					log.Errorw("Failed to read request body for logging", "error", err)
				}
			}

			ct := r.Header.Get("Content-Type")
			ce := r.Header.Get("Content-Encoding")
			next.ServeHTTP(&lw, r) // Hand the request to the actual handler with out wrapper-writer

			duration := time.Since(start)

			log.Infow(
				"request completed",
				"uri", r.RequestURI,
				"method", r.Method,
				"request body", string(bodyHead),
				"request body truncated", bodyTruncated,
				"request content length", r.ContentLength,
				"request content type", ct,
				"request content encoding", ce,
				"response status", responseData.status,
				"duration", duration,
				"response size", responseData.size,
				"response error", responseData.err,
			)
		})
	}
}
