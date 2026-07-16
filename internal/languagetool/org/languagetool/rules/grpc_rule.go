package rules

import (
	"strings"
	"sync"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const (
	GRPCRuleConfigType = "grpc"
	GRPCDefaultBatch   = 8
)

// GRPCRule ports org.languagetool.rules.GRPCRule as RemoteRule with batching.
// Actual gRPC is deferred; Client is a pluggable Match function.
type GRPCRule struct {
	*RemoteRule
	// BatchSize defaults to 8.
	BatchSize int
	// MatchSentences scores a batch of sentences and returns matches per sentence.
	MatchSentences func(sentences []*languagetool.AnalyzedSentence) [][]*RuleMatch
	// MessageBySubID optional message templates.
	MessageBySubID map[string]string
	// Description override.
	Description string
	// Circuit breaker optional.
	Breaker *tools.CircuitBreaker
}

// NewGRPCRule builds a GRPC-backed remote rule surface.
func NewGRPCRule(languageCode string, config *RemoteRuleConfig) *GRPCRule {
	if config == nil {
		config = NewRemoteRuleConfig()
	}
	if config.Type == "" {
		config.Type = GRPCRuleConfigType
	}
	base := NewRemoteRule(languageCode, config)
	g := &GRPCRule{
		RemoteRule: base,
		BatchSize:  GRPCDefaultBatch,
		Breaker:    tools.CircuitBreakerRegistry().GetOrCreate("grpc-rule-" + base.GetID()),
	}
	base.Execute = func(sentences []*languagetool.AnalyzedSentence) *RemoteRuleResult {
		ms := g.run(sentences)
		return &RemoteRuleResult{Matches: ms}
	}
	return g
}

func (g *GRPCRule) run(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	if g == nil || g.MatchSentences == nil {
		return nil
	}
	if g.Breaker != nil && !g.Breaker.Allow() {
		return nil
	}
	batch := g.BatchSize
	if batch <= 0 {
		batch = GRPCDefaultBatch
	}
	var all []*RuleMatch
	for i := 0; i < len(sentences); i += batch {
		end := i + batch
		if end > len(sentences) {
			end = len(sentences)
		}
		chunk := sentences[i:end]
		// optional whitespace normalisation of token text is deferred
		start := time.Now()
		perSent := g.MatchSentences(chunk)
		_ = start
		if g.Breaker != nil {
			g.Breaker.OnSuccess()
		}
		for si, ms := range perSent {
			if si >= len(chunk) {
				break
			}
			for _, m := range ms {
				if m == nil {
					continue
				}
				// attach messages from map when short
				if g.MessageBySubID != nil {
					if r, ok := m.Rule.(interface{ GetID() string }); ok {
						if msg, ok := g.MessageBySubID[r.GetID()]; ok && m.Message == "" {
							m.Message = msg
						}
					}
				}
				all = append(all, m)
			}
		}
	}
	return all
}

// CreateGRPCRule is the Java GRPCRule.create convenience (message map by match id).
func CreateGRPCRule(languageCode string, config *RemoteRuleConfig, ruleID, description string, messages map[string]string) *GRPCRule {
	if config == nil {
		config = NewRemoteRuleConfig()
	}
	if config.RuleID == "" {
		config.RuleID = ruleID
	}
	g := NewGRPCRule(languageCode, config)
	g.Description = description
	g.MessageBySubID = messages
	return g
}

// GRPCPostProcessing ports org.languagetool.rules.GRPCPostProcessing.
// Applies a pluggable transform to matches for a language.
type GRPCPostProcessing struct {
	Config *RemoteRuleConfig
	// Process rewrites matches; nil is identity.
	Process func(sentences []*languagetool.AnalyzedSentence, matches []*RuleMatch) []*RuleMatch
	Breaker *tools.CircuitBreaker
}

const GRPCPostConfigType = "grpc-post"

var (
	grpcPostMu        sync.Mutex
	grpcPostByID      = map[string]*GRPCPostProcessing{}
	grpcPostIDsByLang = map[string]map[string]struct{}{}
)

// ConfigureGRPCPostProcessing registers configs of type grpc-post for a language.
func ConfigureGRPCPostProcessing(langCode string, configs []*RemoteRuleConfig) {
	grpcPostMu.Lock()
	defer grpcPostMu.Unlock()
	if grpcPostIDsByLang[langCode] == nil {
		grpcPostIDsByLang[langCode] = map[string]struct{}{}
	}
	for _, c := range configs {
		if c == nil || !strings.EqualFold(c.Type, GRPCPostConfigType) {
			continue
		}
		id := c.RuleID
		grpcPostIDsByLang[langCode][id] = struct{}{}
		if _, ok := grpcPostByID[id]; ok {
			continue
		}
		grpcPostByID[id] = &GRPCPostProcessing{
			Config:  c,
			Breaker: tools.CircuitBreakerRegistry().GetOrCreate("grpc-post-" + id),
		}
	}
}

// GetGRPCPostProcessing returns configured processors for a language.
func GetGRPCPostProcessing(langCode string) []*GRPCPostProcessing {
	grpcPostMu.Lock()
	defer grpcPostMu.Unlock()
	ids := grpcPostIDsByLang[langCode]
	var out []*GRPCPostProcessing
	for id := range ids {
		if p := grpcPostByID[id]; p != nil {
			out = append(out, p)
		}
	}
	return out
}

// Apply runs Process if allowed by the circuit breaker.
func (p *GRPCPostProcessing) Apply(sentences []*languagetool.AnalyzedSentence, matches []*RuleMatch) []*RuleMatch {
	if p == nil {
		return matches
	}
	if p.Breaker != nil && !p.Breaker.Allow() {
		return matches
	}
	if p.Process == nil {
		return matches
	}
	out := p.Process(sentences, matches)
	if p.Breaker != nil {
		p.Breaker.OnSuccess()
	}
	return out
}

// ResetGRPCPostProcessing clears the registry (tests).
func ResetGRPCPostProcessing() {
	grpcPostMu.Lock()
	defer grpcPostMu.Unlock()
	grpcPostByID = map[string]*GRPCPostProcessing{}
	grpcPostIDsByLang = map[string]map[string]struct{}{}
}
