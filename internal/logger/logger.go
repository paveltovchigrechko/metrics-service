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

			var bodyBytes []byte
			if r.Body != nil {
				var err error
				bodyBytes, err = io.ReadAll(r.Body)
				if err != nil {
					log.Errorw("Failed to read request body for logging", "error", err)
				}
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			ct := r.Header.Get("Content-Type")
			ce := r.Header.Get("Content-Encoding")
			next.ServeHTTP(&lw, r) // Hand the request to the actual handler with out wrapper-writer

			duration := time.Since(start)

			log.Infow(
				"request completed",
				"uri", r.RequestURI,
				"method", r.Method,
				"request body", string(bodyBytes),
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
