package server

import (
	"net/url"
	"strings"
)

// LanguageToolHttpHandler ports org.languagetool.server.LanguageToolHttpHandler routing surface.
type LanguageToolHttpHandler struct {
	Config              *HTTPServerConfig
	AllowedIPs          map[string]struct{}
	RequestLimiter      *RequestLimiter
	ErrorRequestLimiter *ErrorRequestLimiter
	Server              *Server
	TextCheckerV2       *V2TextChecker
	ApiV2               *ApiV2
	ReqCounter          *RequestCounter
	shutdown            bool
}

func NewLanguageToolHttpHandler(
	cfg *HTTPServerConfig,
	allowed map[string]struct{},
	internal bool,
	reqLimiter *RequestLimiter,
	errLimiter *ErrorRequestLimiter,
	srv *Server,
) *LanguageToolHttpHandler {
	if cfg == nil {
		cfg = NewHTTPServerConfig()
	}
	rc := NewRequestCounter()
	return &LanguageToolHttpHandler{
		Config:              cfg,
		AllowedIPs:          allowed,
		RequestLimiter:      reqLimiter,
		ErrorRequestLimiter: errLimiter,
		Server:              srv,
		TextCheckerV2:       NewV2TextChecker(cfg, internal, rc),
		ApiV2:               NewApiV2(cfg, nil),
		ReqCounter:          rc,
	}
}

func (h *LanguageToolHttpHandler) Shutdown() {
	if h != nil {
		h.shutdown = true
	}
}

func (h *LanguageToolHttpHandler) IsShutdown() bool {
	return h != nil && h.shutdown
}

// HandlePath routes a path + query without a real HTTP exchange.
// path is like "/v2/check" or "/v2/languages".
func (h *LanguageToolHttpHandler) HandlePath(path, remoteIP string, query url.Values) (HandleResult, error) {
	if h == nil || h.shutdown {
		return HandleResult{}, NewUnavailableError("handler shutdown", nil)
	}
	if h.AllowedIPs != nil {
		if _, ok := h.AllowedIPs[remoteIP]; !ok && remoteIP != "" {
			return HandleResult{Status: 403, Body: "IP not allowed"}, nil
		}
	}
	reqID := h.ReqCounter.IncrementRequestCount()
	h.ReqCounter.IncrementHandleCount(remoteIP, reqID)
	defer h.ReqCounter.DecrementHandleCount(reqID)

	if h.RequestLimiter != nil && remoteIP != "" && !h.RequestLimiter.Allow(remoteIP) {
		Metrics().LogRequestError(RequestErrorQueueFull)
		return HandleResult{}, NewTooManyRequestsError("Request limit exceeded")
	}

	params := map[string]string{}
	for k, vs := range query {
		if len(vs) > 0 {
			params[k] = vs[0]
		}
	}

	p := strings.TrimPrefix(path, "/")
	if strings.HasPrefix(p, "v2/") || p == "v2" {
		return h.ApiV2.Handle(strings.TrimPrefix(p, "v2/"), params)
	}
	// legacy /check alias → v2
	if p == "check" || strings.HasPrefix(p, "check?") {
		return h.ApiV2.Handle("check", params)
	}
	return HandleResult{}, NewPathNotFoundError("Unsupported path: " + path)
}
