package server

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ServeHTTP implements http.Handler for LanguageToolHttpHandler.
// Soft wire-up: routes /v2/* and legacy /check through HandlePath.
func (h *LanguageToolHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h == nil {
		http.Error(w, "handler not initialized", http.StatusServiceUnavailable)
		return
	}
	if r == nil {
		http.Error(w, "nil request", http.StatusBadRequest)
		return
	}
	// Soft CORS: always set Allow-Origin when configured; handle OPTIONS preflight.
	if h.Config != nil && h.Config.AllowOriginURL != "" {
		w.Header().Set("Access-Control-Allow-Origin", h.Config.AllowOriginURL)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
	}
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	start := time.Now()
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = newRequestID()
	}
	w.Header().Set("X-Request-ID", reqID)

	remoteIP := clientIP(r)
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = r.Header.Get("Origin")
	}

	// merge query + form POST fields (text/data often POSTed)
	q := r.URL.Query()
	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		for k, vs := range r.Form {
			if len(vs) > 0 && q.Get(k) == "" {
				q.Set(k, vs[0])
			}
		}
	}

	path := r.URL.Path
	if path == "" {
		path = "/"
	}
	res, err := h.HandlePathWithReferrer(path, remoteIP, referer, q)
	if err != nil {
		writeHandlerError(w, err)
		return
	}
	status := res.Status
	if status == 0 {
		status = http.StatusOK
	}
	ct := res.ContentType
	if ct == "" {
		ct = "text/plain; charset=utf-8"
	}
	w.Header().Set("Content-Type", ct)
	// Soft discovery headers for clients/proxies.
	w.Header().Set("X-LanguageTool-Software", "LanguageTool-Go")
	w.Header().Set("X-LanguageTool-API-Version", "1")
	w.Header().Set("X-LanguageTool-Time-ms", strconv.FormatInt(time.Since(start).Milliseconds(), 10))
	w.WriteHeader(status)
	_, _ = io.WriteString(w, res.Body)
}

func newRequestID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(b[:])
}

func clientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return tools.JavaStringTrim(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func writeHandlerError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	status := http.StatusInternalServerError
	msg := err.Error()
	switch {
	case errors.As(err, new(*BadRequestError)):
		status = http.StatusBadRequest
	case errors.As(err, new(*TextTooLongError)):
		status = http.StatusRequestEntityTooLarge
	case errors.As(err, new(*TooManyRequestsError)):
		status = http.StatusTooManyRequests
	case errors.As(err, new(*AuthError)):
		status = http.StatusForbidden
	case errors.As(err, new(*PathNotFoundError)):
		status = http.StatusNotFound
	case errors.As(err, new(*UnavailableError)):
		status = http.StatusServiceUnavailable
	}
	http.Error(w, msg, status)
}

// ListenAndServe starts a soft HTTP server on cfg host/port (blocking).
func (s *HTTPServer) ListenAndServe() error {
	if s == nil || s.Handler == nil {
		return NewUnavailableError("server not initialized", nil)
	}
	addr := s.Host
	if addr == "" {
		addr = "127.0.0.1"
	}
	port := s.Port
	if port == 0 && s.Config != nil {
		port = s.Config.Port
	}
	if port == 0 {
		port = DefaultPort
	}
	s.Run()
	return http.ListenAndServe(net.JoinHostPort(addr, itoa(port)), s.Handler)
}
