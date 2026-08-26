package server

// HTTPSServer ports org.languagetool.server.HTTPSServer (surface without TLS bind).
type HTTPSServer struct {
	*HTTPServer
	TLSConfig *HTTPSServerConfig
}

func NewHTTPSServer(cfg *HTTPSServerConfig, internal bool, host string, allowed map[string]struct{}) *HTTPSServer {
	if cfg == nil {
		cfg = NewHTTPSServerConfig("", "")
	}
	http := NewHTTPServerConfig2(cfg.HTTPServerConfig, internal, host, allowed)
	return &HTTPSServer{HTTPServer: http, TLSConfig: cfg}
}

func (s *HTTPSServer) Protocol() string { return "https" }

// HasKeystore reports whether a keystore path is configured.
func (s *HTTPSServer) HasKeystore() bool {
	return s != nil && s.TLSConfig != nil && s.TLSConfig.KeystorePath != ""
}
