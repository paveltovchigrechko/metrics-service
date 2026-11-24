package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrustedSubnetMiddleware(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testCases := []struct {
		name     string
		cidr     string
		method   string
		path     string
		realIP   string // empty means the header is not sent
		wantCode int
	}{
		{"IPv4 inside the subnet", "192.168.1.0/24", http.MethodPost, "/updates", "192.168.1.77", http.StatusOK},
		{"first address of the subnet", "192.168.1.0/24", http.MethodPost, "/updates", "192.168.1.0", http.StatusOK},
		{"last address of the subnet", "192.168.1.0/24", http.MethodPost, "/updates", "192.168.1.255", http.StatusOK},
		{"IPv4 outside the subnet", "192.168.1.0/24", http.MethodPost, "/updates", "192.168.2.1", http.StatusForbidden},
		{"header with spaces", "192.168.1.0/24", http.MethodPost, "/updates", " 192.168.1.5 ", http.StatusOK},
		{"missing header", "192.168.1.0/24", http.MethodPost, "/updates", "", http.StatusForbidden},
		{"malformed header", "192.168.1.0/24", http.MethodPost, "/updates", "not-an-ip", http.StatusForbidden},
		{"several addresses are malformed", "192.168.1.0/24", http.MethodPost, "/updates", "192.168.1.5, 10.0.0.1", http.StatusForbidden},
		{"IPv6 inside the subnet", "2001:db8::/32", http.MethodPost, "/updates", "2001:db8::1", http.StatusOK},
		{"IPv6 outside the subnet", "2001:db8::/32", http.MethodPost, "/updates", "2001:dead::1", http.StatusForbidden},
		{"IPv4 against an IPv6 subnet", "2001:db8::/32", http.MethodPost, "/updates", "192.168.1.5", http.StatusForbidden},
		{"IPv4-mapped IPv6 address", "192.168.1.0/24", http.MethodPost, "/updates", "::ffff:192.168.1.9", http.StatusOK},
		{"single-address subnet", "10.1.2.3/32", http.MethodPost, "/updates", "10.1.2.3", http.StatusOK},
		{"empty subnet allows everyone", "", http.MethodPost, "/updates", "8.8.8.8", http.StatusOK},
		{"empty subnet allows a missing header", "", http.MethodPost, "/updates", "", http.StatusOK},
		{"single update is checked", "192.168.1.0/24", http.MethodPost, "/update/", "10.0.0.1", http.StatusForbidden},
		{"url update is checked", "192.168.1.0/24", http.MethodPost, "/update/gauge/Alloc/1", "10.0.0.1", http.StatusForbidden},
		{"reading the main page is not checked", "192.168.1.0/24", http.MethodGet, "/", "10.0.0.1", http.StatusOK},
		{"reading a value is not checked", "192.168.1.0/24", http.MethodPost, "/value/", "", http.StatusOK},
		{"ping is not checked", "192.168.1.0/24", http.MethodGet, "/ping", "", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mw, err := TrustedSubnetMiddleware(tc.cidr)
			require.NoError(t, err)

			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.realIP != "" {
				req.Header.Set(RealIPHeader, tc.realIP)
			}
			rec := httptest.NewRecorder()
			mw(ok).ServeHTTP(rec, req)

			assert.Equal(t, tc.wantCode, rec.Code)
		})
	}
}

func TestTrustedSubnetMiddleware_DoesNotCallHandlerWhenForbidden(t *testing.T) {
	mw, err := TrustedSubnetMiddleware("192.168.1.0/24")
	require.NoError(t, err)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

	req := httptest.NewRequest(http.MethodPost, "/updates", nil)
	req.Header.Set(RealIPHeader, "10.0.0.1")
	rec := httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.False(t, called)
}

func TestTrustedSubnetMiddleware_InvalidCIDR(t *testing.T) {
	for _, cidr := range []string{"192.168.1.1", "192.168.1.0/33", "nope", "192.168.1.0/"} {
		t.Run(cidr, func(t *testing.T) {
			mw, err := TrustedSubnetMiddleware(cidr)
			assert.Error(t, err)
			assert.Nil(t, mw)
		})
	}
}
