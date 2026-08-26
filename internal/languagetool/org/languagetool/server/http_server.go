package server

// HTTPServer ports org.languagetool.server.HTTPServer (wiring without real bind by default).
type HTTPServer struct {
	*Server
	Internal   bool
	AllowedIPs map[string]struct{}
	Handler    *LanguageToolHttpHandler
}

func NewHTTPServer() *HTTPServer {
	return NewHTTPServerConfig2(NewHTTPServerConfig(), false, DefaultHost, DefaultAllowedIPs)
}

func NewHTTPServerWithConfig(cfg *HTTPServerConfig) *HTTPServer {
	return NewHTTPServerConfig2(cfg, false, DefaultHost, DefaultAllowedIPs)
}

func NewHTTPServerConfig2(cfg *HTTPServerConfig, internal bool, host string, allowed map[string]struct{}) *HTTPServer {
	if cfg == nil {
		cfg = NewHTTPServerConfig()
	}
	base := NewServer(cfg)
	if host != "" {
		base.Host = host
	}
	if allowed == nil && !cfg.PublicAccess {
		allowed = DefaultAllowedIPs
	}
	h := &HTTPServer{
		Server:     base,
		Internal:   internal,
		AllowedIPs: allowed,
	}
	h.Handler = NewLanguageToolHttpHandler(cfg, allowed, internal, base.RequestLimiter, base.ErrorRequestLimiter, base)
	return h
}

func (s *HTTPServer) Protocol() string { return "http" }

// AllowIP checks remote IP against the allow list (nil list = allow all).
func (s *HTTPServer) AllowIP(ip string) bool {
	if s == nil || s.AllowedIPs == nil {
		return true
	}
	_, ok := s.AllowedIPs[ip]
	return ok
}
