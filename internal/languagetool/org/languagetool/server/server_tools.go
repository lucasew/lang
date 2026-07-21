package server

import (
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// GetHTTPRequestIp ports ServerTools.getHTTPRequestIp-like extraction.
func GetHTTPRequestIP(r *http.Request, trustXForwardedFor bool) string {
	if r == nil {
		return ""
	}
	if trustXForwardedFor {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			// Typical Java XFF first hop trim (String.trim).
			return tools.JavaStringTrim(parts[0])
		}
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return tools.JavaStringTrim(xri)
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
	q = tools.JavaStringTrim(q)
	if len(q) > max {
		return q[:max] + "…"
	}
	return q
}

// sentContentRE matches <sentcontent>…</sentcontent> including newlines (Java DOTALL).
var sentContentRE = regexp.MustCompile(`(?s)<sentcontent>.*?</sentcontent>`)

// CleanUserTextFromMessage ports ServerTools.cleanUserTextFromMessage.
// When logging map has inputLogging=no, strips <sentcontent>…</sentcontent> payloads.
func CleanUserTextFromMessage(message string, logging map[string]string) string {
	if logging != nil && strings.EqualFold(logging["inputLogging"], "no") {
		return sentContentRE.ReplaceAllString(message, "<< content removed >>")
	}
	return message
}
