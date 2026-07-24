package server

// Server ports org.languagetool.server.Server — base for HTTP/HTTPS servers.
type Server struct {
	Port    int
	Host    string
	running bool

	RequestLimiter      *RequestLimiter
	ErrorRequestLimiter *ErrorRequestLimiter
	RequestCounter      *RequestCounter
	Config              *HTTPServerConfig
}

// DefaultAllowedIPs ports Server.DEFAULT_ALLOWED_IPS.
var DefaultAllowedIPs = map[string]struct{}{
	"0:0:0:0:0:0:0:1":    {},
	"0:0:0:0:0:0:0:1%0": {},
	"127.0.0.1":          {},
}

func NewServer(cfg *HTTPServerConfig) *Server {
	if cfg == nil {
		cfg = NewHTTPServerConfig()
	}
	s := &Server{
		Port:           cfg.Port,
		Host:           DefaultHost,
		Config:         cfg,
		RequestCounter: NewRequestCounter(),
	}
	s.RequestLimiter = RequestLimiterFromConfig(cfg)
	s.ErrorRequestLimiter = ErrorRequestLimiterFromConfig(cfg)
	return s
}

// RequestLimiterFromConfig builds a limiter when request limits are configured.
func RequestLimiterFromConfig(cfg *HTTPServerConfig) *RequestLimiter {
	if cfg == nil {
		return nil
	}
	if (cfg.RequestLimit > 0 || cfg.RequestLimitInBytes > 0) && cfg.RequestLimitPeriodInSeconds > 0 {
		return NewRequestLimiter(cfg.RequestLimit, cfg.RequestLimitPeriodInSeconds)
	}
	return nil
}

// ErrorRequestLimiterFromConfig builds an error limiter when timeout limits are set.
func ErrorRequestLimiterFromConfig(cfg *HTTPServerConfig) *ErrorRequestLimiter {
	if cfg == nil {
		return nil
	}
	if cfg.TimeoutRequestLimit > 0 && cfg.RequestLimitPeriodInSeconds > 0 {
		return NewErrorRequestLimiter(cfg.TimeoutRequestLimit, cfg.RequestLimitPeriodInSeconds)
	}
	return nil
}

func (s *Server) Protocol() string { return "http" }

// Run marks the server as running (wire real listener later).
func (s *Server) Run() {
	if s == nil {
		return
	}
	s.running = true
}

// Stop marks the server as stopped.
func (s *Server) Stop() {
	if s == nil {
		return
	}
	s.running = false
}

func (s *Server) IsRunning() bool {
	return s != nil && s.running
}

// IsAllowedIP reports whether ip is in the default allow list (non-public mode).
func IsAllowedIP(ip string) bool {
	_, ok := DefaultAllowedIPs[ip]
	return ok
}

// UsageRequested reports if CLI args request help.
func UsageRequested(args []string) bool {
	return len(args) == 1 && (args[0] == "-h" || args[0] == "--help")
}
