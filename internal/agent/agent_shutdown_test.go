package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRun_SendsInFlightMetricsAfterCancel checks that cancelling the agent
// does not drop a request that is already in progress or a batch queued
// behind it. Workers keep sending until the reporter has stopped and the
// queue is empty.
func TestRun_SendsInFlightMetricsAfterCancel(t *testing.T) {
	var requests atomic.Int32
	firstSeen := make(chan struct{})
	release := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 1 {
			close(firstSeen)
			<-release
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	a := newTestAgent(t, "-l", "1", "-a", strings.TrimPrefix(server.URL, "http://"))
	a.cfg.ReportInterval = 10 * time.Millisecond
	a.cfg.PollInterval = time.Hour

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		a.Run(ctx)
		close(done)
	}()

	select {
	case <-firstSeen:
	case <-time.After(2 * time.Second):
		t.Fatal("agent did not send the first batch")
	}

	// Let the reporter queue another batch while the first request is held.
	time.Sleep(80 * time.Millisecond)
	cancel()

	select {
	case <-done:
		t.Fatal("agent returned before the in-flight request finished")
	case <-time.After(50 * time.Millisecond):
	}

	close(release)

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("agent did not return after the in-flight request finished")
	}

	assert.GreaterOrEqual(t, requests.Load(), int32(2))
}
