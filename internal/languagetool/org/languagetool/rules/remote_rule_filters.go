package rules

import (
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
		re  *regexp.Regexp
		pos map[MatchPosition]struct{}
	}
	var blocks []blocked
	for _, fr := range filters {
		if fr == nil || fr.IDPattern == nil {
			continue
		}
		pos := map[MatchPosition]struct{}{}
		if fr.MatchPositions != nil && sentence != nil {
			for _, p := range fr.MatchPositions(sentence) {
				pos[p] = struct{}{}
			}
		}
		blocks = append(blocks, blocked{re: fr.IDPattern, pos: pos})
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
			if !b.re.MatchString(id) {
				continue
			}
			// if filter has no positions, drop all matching ids (aggressive test mode)
			if len(b.pos) == 0 {
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
