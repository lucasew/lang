package rules

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

const RemoteRuleFilterFile = "remote-rule-filters.xml"

// FilterRule is a simplified filter: if it matches the same span as a remote match
// whose rule ID matches IDPattern, the remote match is discarded.
type FilterRule struct {
	// IDPattern is a regex over remote match rule IDs.
	IDPattern *regexp.Regexp
	// MatchPositions returns spans that filter matching remote rule IDs.
	MatchPositions func(sentence *languagetool.AnalyzedSentence) []MatchPosition
}

// RemoteRuleFilters ports org.languagetool.rules.RemoteRuleFilters.
// Pattern XML loading is deferred; filters are registered per language code.
type RemoteRuleFilters struct {
	mu      sync.RWMutex
	byLang  map[string][]*FilterRule
}

// GlobalRemoteRuleFilters is the process-wide registry.
var GlobalRemoteRuleFilters = NewRemoteRuleFilters()

func NewRemoteRuleFilters() *RemoteRuleFilters {
	return &RemoteRuleFilters{byLang: map[string][]*FilterRule{}}
}

// Register adds filters for a language short code.
func (f *RemoteRuleFilters) Register(langCode string, filters ...*FilterRule) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byLang[langCode] = append(f.byLang[langCode], filters...)
}

// Clear removes all filters (tests).
func (f *RemoteRuleFilters) Clear() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byLang = map[string][]*FilterRule{}
}

// GetFilename ports RemoteRuleFilters.getFilename.
func GetRemoteRuleFilterFilename(langShortCode string) string {
	if langShortCode == "de-DE-x-simple-language" {
		langShortCode = "de"
	}
	// strip variant for path: en-US → en
	if i := indexByte(langShortCode, '-'); i > 0 {
		// keep de-DE style as short code only for file path like Java shortCode
		// Java uses lang.getShortCode() which is "en" for en-US
		langShortCode = langShortCode[:i]
	}
	return langShortCode + "/" + RemoteRuleFilterFile
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// FilterMatches discards remote matches that share a span with an active filter rule.
func (f *RemoteRuleFilters) FilterMatches(langCode string, sentence *languagetool.AnalyzedSentence, matches []*RuleMatch) []*RuleMatch {
	if f == nil || len(matches) == 0 {
		return matches
	}
	f.mu.RLock()
	filters := f.byLang[langCode]
	// also try short code
	if len(filters) == 0 {
		if i := indexByte(langCode, '-'); i > 0 {
			filters = f.byLang[langCode[:i]]
		}
	}
	f.mu.RUnlock()
	if len(filters) == 0 {
		return matches
	}

	// build blocked positions per filter id-pattern
	type blocked struct {
		re *regexp.Regexp
		// pos is populated when MatchPositions is set (XML-backed / position filters).
		pos map[MatchPosition]struct{}
		// requirePos: true when MatchPositions is non-nil — only drop on equal span.
		// false (MatchPositions nil): drop every remote id matching the regex (manual tests).
		requirePos bool
	}
	var blocks []blocked
	for _, fr := range filters {
		if fr == nil || fr.IDPattern == nil {
			continue
		}
		b := blocked{re: fr.IDPattern, pos: map[MatchPosition]struct{}{}}
		if fr.MatchPositions != nil {
			b.requirePos = true
			if sentence != nil {
				for _, p := range fr.MatchPositions(sentence) {
					b.pos[p] = struct{}{}
				}
			}
		}
		blocks = append(blocks, b)
	}

	var out []*RuleMatch
	for _, m := range matches {
		if m == nil {
			continue
		}
		id := ruleIDOfMatch(m)
		span := MatchPosition{Start: m.FromPos, End: m.ToPos}
		drop := false
		for _, b := range blocks {
			// Java Pattern.matcher(id).matches() / String.matches — whole-string match.
			if !remoteRuleIDFullMatch(b.re, id) {
				continue
			}
			if !b.requirePos {
				// MatchPositions nil: drop all matching ids (test / aggressive register).
				drop = true
				break
			}
			if _, ok := b.pos[span]; ok {
				drop = true
				break
			}
		}
		if !drop {
			out = append(out, m)
		}
	}
	return out
}

// FilterMatches is a package-level convenience using GlobalRemoteRuleFilters.
func FilterRemoteRuleMatches(langCode string, sentence *languagetool.AnalyzedSentence, matches []*RuleMatch) []*RuleMatch {
	return GlobalRemoteRuleFilters.FilterMatches(langCode, sentence, matches)
}

func ruleIDOfMatch(m *RuleMatch) string {
	if m == nil || m.Rule == nil {
		return ""
	}
	if r, ok := m.Rule.(interface{ GetID() string }); ok {
		return r.GetID()
	}
	return ""
}

// remoteRuleIDFullMatch ports Java Matcher.matches() (entire region must match).
func remoteRuleIDFullMatch(re *regexp.Regexp, id string) bool {
	if re == nil {
		return false
	}
	loc := re.FindStringIndex(id)
	return loc != nil && loc[0] == 0 && loc[1] == len(id)
}

// CompileRemoteRuleIDPattern compiles a filter rule id as a regex over remote match IDs
// (Java RemoteRuleFilters.compilePatterns).
func CompileRemoteRuleIDPattern(ruleID string) (*regexp.Regexp, error) {
	if ruleID == "" {
		return nil, fmt.Errorf("empty remote rule filter id")
	}
	return regexp.Compile(ruleID)
}
