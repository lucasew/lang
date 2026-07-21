package server

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AbTestService ports org.languagetool.server.AbTestService.
type AbTestService interface {
	GetActiveAbTestForClient(params map[string]string, cfg *HTTPServerConfig) []string
}

// LocalAbTestService ports org.languagetool.server.LocalAbTestService.
type LocalAbTestService struct {
	// Allowed lists experiments configured on the server.
	Allowed []string
	// Clients matches useragent when non-nil.
	Clients *regexp.Regexp
}

func NewLocalAbTestService(allowed []string, clients *regexp.Regexp) *LocalAbTestService {
	return &LocalAbTestService{Allowed: allowed, Clients: clients}
}

func (s *LocalAbTestService) GetActiveAbTestForClient(params map[string]string, cfg *HTTPServerConfig) []string {
	if s == nil || params == nil {
		return nil
	}
	agent := params["useragent"]
	if agent == "" {
		agent = "unknown"
	}
	paramActivated := params["abtest"]
	if paramActivated == "" {
		return nil
	}
	if s.Clients != nil && !s.Clients.MatchString(agent) {
		return nil
	}
	allowed := map[string]struct{}{}
	for _, a := range s.Allowed {
		allowed[a] = struct{}{}
	}
	if cfg != nil {
		// config can carry ab tests via DisabledRuleIDs-like list; use Allowed primarily
	}
	// Java LocalAbTestService: paramActivatedAbTest.trim().split(",") then abParam.trim().
	var out []string
	for _, p := range strings.Split(tools.JavaStringTrim(paramActivated), ",") {
		p = tools.JavaStringTrim(p)
		if p == "" {
			continue
		}
		if _, ok := allowed[p]; ok {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// DefaultAbTestService returns the open-source local service.
func DefaultAbTestService() AbTestService {
	return NewLocalAbTestService(nil, nil)
}
