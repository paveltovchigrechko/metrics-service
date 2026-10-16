package logger

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggerMiddleware(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	log := zap.New(core).Sugar()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte("created"))
		require.NoError(t, err)
	})

	handler := LoggerMiddleware(log)(next)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test/10", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, "created", recorder.Body.String())

	require.Equal(t, 1, logs.Len())

	entry := logs.All()[0]
	assert.Equal(t, "request completed", entry.Message)

	fields := entry.ContextMap()
	assert.Equal(t, http.MethodPost, fields["method"])
	assert.Equal(t, "/update/gauge/test/10", fields["uri"])
	assert.Equal(t, int64(http.StatusCreated), fields["response status"])
	assert.Equal(t, int64(len("created")), fields["response size"])
	assert.Contains(t, fields, "duration")
}

func TestLoggerMiddleware_RequestBody(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		wantLogged    string
		wantTruncated bool
	}{
		{name: "empty body", body: "", wantLogged: "", wantTruncated: false},
		{name: "short body is logged completely", body: `{"id":"Alloc"}`, wantLogged: `{"id":"Alloc"}`, wantTruncated: false},
		{
			name:          "body of exactly the limit is not truncated",
			body:          strings.Repeat("a", MaxLoggedBodySize),
			wantLogged:    strings.Repeat("a", MaxLoggedBodySize),
			wantTruncated: false,
		},
		{
			name:          "long body is cut to the limit",
			body:          strings.Repeat("a", MaxLoggedBodySize+1),
			wantLogged:    strings.Repeat("a", MaxLoggedBodySize),
			wantTruncated: true,
		},
		{
			name:          "very long body is cut to the limit",
			body:          strings.Repeat("b", 100*1024),
			wantLogged:    strings.Repeat("b", MaxLoggedBodySize),
			wantTruncated: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			core, logs := observer.New(zap.InfoLevel)
			log := zap.New(core).Sugar()

			var received string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				received = string(data)
			})

			req := httptest.NewRequest(http.MethodPost, "/updates", strings.NewReader(tc.body))
			LoggerMiddleware(log)(next).ServeHTTP(httptest.NewRecorder(), req)

			// The handler must still get the whole body, whatever is logged.
			assert.Equal(t, tc.body, received)

			require.Equal(t, 1, logs.Len())
			fields := logs.All()[0].ContextMap()
			assert.Equal(t, tc.wantLogged, fields["request body"])
			assert.Equal(t, tc.wantTruncated, fields["request body truncated"])
			assert.Equal(t, int64(len(tc.body)), fields["request content length"])
		})
	}
}
