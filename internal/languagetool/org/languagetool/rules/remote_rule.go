package rules

import (
	"regexp"
	"strconv"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// RemoteRule ports org.languagetool.rules.RemoteRule configuration surface
// (actual RPC deferred; Execute is pluggable).
type RemoteRule struct {
	Config                   *RemoteRuleConfig
	LanguageCode             string
	Premium                  bool
	InputLogging             bool
	FilterMatches            bool
	FixOffsets               bool
	WhitespaceNormalisation  bool
	IncludedInHiddenMatches  bool
	SuppressMisspelledMatch  *regexp.Regexp
	SuppressMisspelledSugg   *regexp.Regexp
	// Execute is the backend call; if nil, Match returns empty.
	Execute func(sentences []*languagetool.AnalyzedSentence) *RemoteRuleResult
}

func NewRemoteRule(languageCode string, config *RemoteRuleConfig) *RemoteRule {
	if config == nil {
		config = NewRemoteRuleConfig()
	}
	r := &RemoteRule{
		Config:                  config,
		LanguageCode:            languageCode,
		InputLogging:            true,
		WhitespaceNormalisation: true,
		FixOffsets:              true,
		IncludedInHiddenMatches: true,
	}
	if config.Options != nil {
		r.FilterMatches = boolOpt(config.Options, "filterMatches", false)
		r.WhitespaceNormalisation = boolOpt(config.Options, "whitespaceNormalisation", true)
		r.FixOffsets = boolOpt(config.Options, "fixOffsets", true)
		r.Premium = boolOpt(config.Options, "premium", false)
		r.IncludedInHiddenMatches = boolOpt(config.Options, "includedInHiddenMatches", true)
		if v, ok := config.Options["suppressMisspelledMatch"]; ok && v != "" {
			r.SuppressMisspelledMatch = regexp.MustCompile(v)
		}
		if v, ok := config.Options["suppressMisspelledSuggestions"]; ok && v != "" {
			r.SuppressMisspelledSugg = regexp.MustCompile(v)
		}
	}
	return r
}

func (r *RemoteRule) GetID() string {
	if r.Config != nil && r.Config.RuleID != "" {
		return r.Config.RuleID
	}
	return "REMOTE_RULE"
}

func (r *RemoteRule) GetDescription() string { return "Remote rule" }

// MatchRemote runs Execute over sentences and returns matches.
func (r *RemoteRule) MatchRemote(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	if r == nil || r.Execute == nil {
		return nil
	}
	res := r.Execute(sentences)
	if res == nil {
		return nil
	}
	return res.Matches
}

func boolOpt(m map[string]string, key string, def bool) bool {
	v, ok := m[key]
	if !ok {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}
