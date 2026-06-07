// Package httputil contains small helpers shared by HTTP-facing services.
package httputil

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP resolves the caller IP from the request.
//
// Why this helper exists:
//   - legacy handlers use the caller IP in downstream RPC requests
//   - deployments may sit behind a reverse proxy
//   - repeating the same resolution code in every handler would be noisy and
//     easy to get wrong
func ClientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return normalizeLoopback(ip)
	}

	if ip := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); ip != "" {
		parts := strings.Split(ip, ",")
		if len(parts) > 0 {
			return normalizeLoopback(strings.TrimSpace(parts[0]))
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return normalizeLoopback(r.RemoteAddr)
	}

	return normalizeLoopback(host)
}

func normalizeLoopback(ip string) string {
	if ip == "::1" {
		return "127.0.0.1"
	}
	return ip
}
