package server

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
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
	if h.Config != nil && h.Config.AllowOriginURL != "" {
		w.Header().Set("Access-Control-Allow-Origin", h.Config.AllowOriginURL)
	}
	w.WriteHeader(status)
	_, _ = io.WriteString(w, res.Body)
}

func clientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
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
