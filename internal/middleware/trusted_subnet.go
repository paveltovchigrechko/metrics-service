package middleware

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/paveltovchigrechko/metrics-service/internal/handler"
)

// RealIPHeader is the request header in which the agent reports its address.
const RealIPHeader = "X-Real-IP"

// updatePathPrefix matches the endpoints that accept metrics from agents:
// /update/..., /update and /updates.
const updatePathPrefix = "/update"

var errUntrustedIP = errors.New("agent address is not in the trusted subnet")

// TrustedSubnetMiddleware accepts metric updates only from agents whose
// RealIPHeader address is inside the CIDR subnet. Other requests get 403
// Forbidden, including requests with a missing or malformed header.
//
// The header is set by the client, so use this for filtering by deployment
// network, not as authentication.
func TrustedSubnetMiddleware(cidr string) (func(http.Handler) http.Handler, error) {
	cidr = strings.TrimSpace(cidr)
	if cidr == "" {
		return func(next http.Handler) http.Handler {
			return next
		}, nil
	}

	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("trusted subnet %q: %w", cidr, err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if the request is for an update endpoint.
			if !strings.HasPrefix(r.URL.Path, updatePathPrefix) {
				next.ServeHTTP(w, r)
				return
			}

			ip := net.ParseIP(strings.TrimSpace(r.Header.Get(RealIPHeader)))
			if ip == nil || !subnet.Contains(ip) {
				handler.WriteError(w, errUntrustedIP, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}, nil
}
