package audit

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testEvent = &Event{TS: 123, Metrics: []string{"Alloc", "Frees"}, IP: "192.168.0.42"}

func TestFileObserver_AppendsLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	o, err := NewFileObserver(path)
	require.NoError(t, err)

	require.NoError(t, o.WriteEvent(testEvent))
	require.NoError(t, o.WriteEvent(testEvent))
	require.NoError(t, o.Close())

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	require.Len(t, lines, 2)
	for _, l := range lines {
		var got Event
		require.NoError(t, json.Unmarshal([]byte(l), &got))
		assert.Equal(t, *testEvent, got)
	}
}

func TestFileObserver_AppendsToExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	require.NoError(t, os.WriteFile(path, []byte("existing\n"), 0600))

	o, err := NewFileObserver(path)
	require.NoError(t, err)
	require.NoError(t, o.WriteEvent(testEvent))
	require.NoError(t, o.Close())

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(data), "existing\n"))
}

func TestFileObserver_FailsOnBadPath(t *testing.T) {
	_, err := NewFileObserver(filepath.Join(t.TempDir(), "no-such-dir", "audit.log"))
	assert.Error(t, err)
}

func TestFileObserver_WriteAfterClose(t *testing.T) {
	o, err := NewFileObserver(filepath.Join(t.TempDir(), "audit.log"))
	require.NoError(t, err)
	require.NoError(t, o.Close())

	assert.Error(t, o.WriteEvent(testEvent))
}

func TestHTTPObserver(t *testing.T) {
	t.Run("posts event", func(t *testing.T) {
		var got Event
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		}))
		defer srv.Close()

		require.NoError(t, NewHTTPObserver(srv.URL).WriteEvent(testEvent))
		assert.Equal(t, *testEvent, got)
	})

	t.Run("non-2xx is an error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		assert.Error(t, NewHTTPObserver(srv.URL).WriteEvent(testEvent))
	})
}

type fakeObserver struct {
	err   error
	calls int
}

func (f *fakeObserver) WriteEvent(*Event) error {
	f.calls++
	return f.err
}

func TestPublisher_NotifiesAllDespiteErrors(t *testing.T) {
	failing := &fakeObserver{err: errors.New("boom")}
	ok := &fakeObserver{}

	err := NewPublisher(failing, ok).Notify(testEvent)

	assert.Error(t, err)
	assert.Equal(t, 1, failing.calls)
	assert.Equal(t, 1, ok.calls)
}

type blockingObserver struct {
	release chan struct{}
	mu      sync.Mutex
	got     []*Event
}

func (b *blockingObserver) WriteEvent(e *Event) error {
	<-b.release
	b.mu.Lock()
	defer b.mu.Unlock()
	b.got = append(b.got, e)
	return nil
}

func TestAsyncObserver(t *testing.T) {
	t.Run("does not block and delivers on close", func(t *testing.T) {
		inner := &blockingObserver{release: make(chan struct{})}
		a := NewAsyncObserver(inner, 10)

		// The inner observer is blocked, yet WriteEvent returns immediately.
		require.NoError(t, a.WriteEvent(testEvent))
		require.NoError(t, a.WriteEvent(testEvent))

		close(inner.release)
		require.NoError(t, a.Close())

		inner.mu.Lock()
		defer inner.mu.Unlock()
		assert.Len(t, inner.got, 2)
	})

	t.Run("drops events when queue is full", func(t *testing.T) {
		inner := &blockingObserver{release: make(chan struct{})}
		a := NewAsyncObserver(inner, 1)

		var full int
		for i := 0; i < 5; i++ { // 1 may be taken by the worker, 1 sits in the queue.
			if errors.Is(a.WriteEvent(testEvent), ErrQueueFull) {
				full++
			}
		}
		assert.GreaterOrEqual(t, full, 3)

		close(inner.release)
		require.NoError(t, a.Close())
	})

	t.Run("rejects events after close", func(t *testing.T) {
		a := NewAsyncObserver(&fakeObserver{}, 1)
		require.NoError(t, a.Close())
		assert.Error(t, a.WriteEvent(testEvent))
		assert.NoError(t, a.Close()) // Idempotent.
	})

	t.Run("closes the wrapped observer once after delivering events", func(t *testing.T) {
		inner := &closingObserver{}
		a := NewAsyncObserver(inner, 10)
		require.NoError(t, a.WriteEvent(testEvent))

		require.NoError(t, a.Close())
		require.NoError(t, a.Close())

		assert.Equal(t, 1, inner.closes)
		assert.True(t, inner.wroteBeforeClose)
	})

	t.Run("returns the close error of the wrapped observer", func(t *testing.T) {
		a := NewAsyncObserver(&closingObserver{closeErr: errors.New("close failed")}, 1)

		assert.Error(t, a.Close())
		assert.Error(t, a.Close()) // The first result is kept.
	})
}

type closingObserver struct {
	writes           int
	closes           int
	wroteBeforeClose bool
	closeErr         error
}

func (c *closingObserver) WriteEvent(*Event) error {
	c.writes++
	return nil
}

func (c *closingObserver) Close() error {
	c.closes++
	c.wroteBeforeClose = c.writes > 0
	return c.closeErr
}
