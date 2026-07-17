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
		ApiV2:               NewApiV2(cfg, DefaultCoreLanguages()),
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
	return h.HandlePathWithReferrer(path, remoteIP, "", query)
}

// HandlePathWithReferrer is HandlePath with an optional HTTP Referer / Origin value.
func (h *LanguageToolHttpHandler) HandlePathWithReferrer(path, remoteIP, referer string, query url.Values) (HandleResult, error) {
	if h == nil || h.shutdown {
		return HandleResult{}, NewUnavailableError("handler shutdown", nil)
	}
	if h.AllowedIPs != nil {
		if _, ok := h.AllowedIPs[remoteIP]; !ok && remoteIP != "" {
			return HandleResult{Status: 403, Body: "IP not allowed"}, nil
		}
	}
	if h.Config != nil && h.Config.IsBlockedReferrer(referer) {
		return HandleResult{Status: 403, Body: "Referrer not allowed"}, nil
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
	// soft API index for discovery (root and bare /v2)
	if p == "" || p == "v2" || p == "v2/" {
		body := `{"name":"LanguageTool-Go","apiVersion":1,"endpoints":["/v2/check","/v2/languages","/v2/info","/v2/configinfo","/v2/maxtextlength","/v2/words","/v2/words/add","/v2/words/delete"]}`
		return HandleResult{Status: 200, ContentType: "application/json", Body: body}, nil
	}
	if strings.HasPrefix(p, "v2/") {
		return h.ApiV2.Handle(strings.TrimPrefix(p, "v2/"), params)
	}
	// legacy /check alias → v2
	if p == "check" || strings.HasPrefix(p, "check?") {
		return h.ApiV2.Handle("check", params)
	}
	return HandleResult{}, NewPathNotFoundError("Unsupported path: " + path)
}
