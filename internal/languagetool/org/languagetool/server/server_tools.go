package server

import (
	"net"
	"net/http"
	"strings"
)

// GetHTTPRequestIp ports ServerTools.getHTTPRequestIp-like extraction.
func GetHTTPRequestIP(r *http.Request, trustXForwardedFor bool) string {
	if r == nil {
		return ""
	}
	if trustXForwardedFor {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			return strings.TrimSpace(parts[0])
		}
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return strings.TrimSpace(xri)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// CleanUserQuery soft-sanitizes user query text for logs (truncate).
func CleanUserQuery(q string, max int) string {
	if max <= 0 {
		max = 200
	}
	q = strings.ReplaceAll(q, "\n", " ")
	q = strings.TrimSpace(q)
	if len(q) > max {
		return q[:max] + "…"
	}
	return q
}
