package handler

import (
	"net"
	"net/http"
	"time"

	"github.com/paveltovchigrechko/metrics-service/internal/audit"
)

func createEvent(req *http.Request, metricNames ...string) *audit.Event {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		ip = req.RemoteAddr // No port in the address, use it as is.
	}

	return &audit.Event{
		TS:      time.Now().Unix(),
		Metrics: metricNames,
		IP:      ip,
	}
}

// audit publishes the audit event about successfully processed metrics.
// It returns the notifier error; callers decide how to handle it (audit failures must not affect the response).
func (h *AppHandler) audit(req *http.Request, metricNames ...string) error {
	if h.notifier == nil {
		return nil
	}

	return h.notifier.Notify(createEvent(req, metricNames...))
}
