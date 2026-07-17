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

// SuppressMisspelled ports RemoteRule.suppressMisspelled:
// - suppressMisspelledMatch: drop match if any suggestion is misspelled
// - suppressMisspelledSuggestions: keep only correctly spelled suggestions; drop if empty
// isMisspelled may be nil (treat all suggestions as correctly spelled).
func (r *RemoteRule) SuppressMisspelled(matches []*RuleMatch, isMisspelled func(string) bool) []*RuleMatch {
	if r == nil || len(matches) == 0 {
		return matches
	}
	if isMisspelled == nil {
		isMisspelled = func(string) bool { return false }
	}
	var out []*RuleMatch
	for _, m := range matches {
		if m == nil {
			continue
		}
		id := ""
		if g, ok := m.Rule.(interface{ GetID() string }); ok {
			id = g.GetID()
		}
		sugs := m.GetSuggestedReplacements()
		// match suppression: drop if not all suggestions are correctly spelled
		// Java Pattern.matcher(id).matches() → full-string match
		if regexFullMatch(r.SuppressMisspelledMatch, id) {
			allOK := true
			for _, s := range sugs {
				if isMisspelled(s) {
					allOK = false
					break
				}
			}
			if !allOK {
				continue
			}
		}
		// suggestion filtering
		if regexFullMatch(r.SuppressMisspelledSugg, id) {
			var kept []string
			for _, s := range sugs {
				if !isMisspelled(s) {
					kept = append(kept, s)
				}
			}
			if len(kept) == 0 {
				continue
			}
			m.SetSuggestedReplacements(kept)
		}
		out = append(out, m)
	}
	return out
}

// regexFullMatch mirrors Java Matcher.matches() (entire string).
func regexFullMatch(re *regexp.Regexp, s string) bool {
	if re == nil {
		return false
	}
	loc := re.FindStringIndex(s)
	return loc != nil && loc[0] == 0 && loc[1] == len(s)
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
