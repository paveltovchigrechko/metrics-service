package agent

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paveltovchigrechko/metrics-service/internal/middleware"
	"github.com/paveltovchigrechko/metrics-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostIP(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(server.Close)

	a := newTestAgent(t, "-a", strings.TrimPrefix(server.URL, "http://"))

	ip, err := a.hostIP()

	require.NoError(t, err)
	parsed := net.ParseIP(ip)
	require.NotNil(t, parsed, "hostIP returned %q", ip)
	// The test server listens on loopback, so the route to it uses loopback.
	assert.True(t, parsed.IsLoopback())
}

// TestHostIP_Localhost covers the default server address. The name must give
// the IPv4 loopback address, not ::1, or a 127.0.0.0/8 subnet would refuse it.
func TestHostIP_Localhost(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(server.Close)
	_, port, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	require.NoError(t, err)

	a := newTestAgent(t, "-a", "localhost:"+port)

	ip, err := a.hostIP()

	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1", ip)
}

func TestSendMetricsJSON_RealIP(t *testing.T) {
	metrics := testMetrics(t)

	t.Run("JSON requests carry the host address", func(t *testing.T) {
		addr, last := newCaptureServer(t)
		a := newTestAgent(t, "-a", addr)

		require.NoError(t, a.sendMetricsJSON(metrics, true))

		got := last().header.Get(middleware.RealIPHeader)
		require.NotEmpty(t, got)
		assert.NotNil(t, net.ParseIP(got), "X-Real-IP is %q", got)
	})

	t.Run("URL requests carry the host address", func(t *testing.T) {
		addr, last := newCaptureServer(t)
		a := newTestAgent(t, "-a", addr)

		m, err := model.CreateMetrics("Alloc", model.Gauge, 0, 1)
		require.NoError(t, err)
		require.NoError(t, a.sendMetricsURL([]*model.Metrics{m}))

		assert.NotNil(t, net.ParseIP(last().header.Get(middleware.RealIPHeader)))
	})
}

// TestSendMetricsJSON_TrustedSubnet runs the agent against the server
// middleware: the same address is accepted or refused depending on the subnet.
func TestSendMetricsJSON_TrustedSubnet(t *testing.T) {
	metrics := testMetrics(t)

	newServer := func(t *testing.T, cidr string) (addr string, handled *int) {
		t.Helper()

		count := 0
		mw, err := middleware.TrustedSubnetMiddleware(cidr)
		require.NoError(t, err)
		server := httptest.NewServer(mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count++
		})))
		t.Cleanup(server.Close)

		return strings.TrimPrefix(server.URL, "http://"), &count
	}

	t.Run("accepted from a trusted subnet", func(t *testing.T) {
		addr, handled := newServer(t, "127.0.0.0/8")
		a := newTestAgent(t, "-a", addr)

		require.NoError(t, a.sendMetricsJSON(metrics, true))
		assert.Equal(t, 1, *handled)
	})

	t.Run("refused with 403 outside the subnet", func(t *testing.T) {
		addr, handled := newServer(t, "10.99.0.0/16")
		a := newTestAgent(t, "-a", addr)

		err := a.sendMetricsJSON(metrics, true)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "403")
		assert.Equal(t, 0, *handled)
	})

	t.Run("accepted without restriction", func(t *testing.T) {
		addr, handled := newServer(t, "")
		a := newTestAgent(t, "-a", addr)

		require.NoError(t, a.sendMetricsJSON(metrics, true))
		assert.Equal(t, 1, *handled)
	})
}
