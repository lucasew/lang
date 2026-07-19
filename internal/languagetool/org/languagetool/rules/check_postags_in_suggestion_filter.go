package rules

import (
	"regexp"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// CheckPostagsInSuggestionFilter ports org.languagetool.rules.CheckPostagsInSuggestionFilter.
// TagToken returns POS tags for a single token; nil → fail-closed (suppress).
type CheckPostagsInSuggestionFilter struct {
	TagToken func(token string) []string
}

func NewCheckPostagsInSuggestionFilter(tag func(string) []string) *CheckPostagsInSuggestionFilter {
	return &CheckPostagsInSuggestionFilter{TagToken: tag}
}

var (
	checkPostagsTagMu sync.RWMutex
	defaultCheckPostagsTag func(string) []string
)

// SetDefaultCheckPostagsTagger wires language tagger for CheckPostagsInSuggestionFilter
// (Java: Language.getTagger()).
func SetDefaultCheckPostagsTagger(tag func(string) []string) {
	checkPostagsTagMu.Lock()
	defer checkPostagsTagMu.Unlock()
	defaultCheckPostagsTag = tag
}

// AcceptRuleMatch ports CheckPostagsInSuggestionFilter.acceptRuleMatch.
func (f *CheckPostagsInSuggestionFilter) AcceptRuleMatch(match *RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	if match == nil {
		return nil
	}
	tag := f.TagToken
	if tag == nil {
		checkPostagsTagMu.RLock()
		tag = defaultCheckPostagsTag
		checkPostagsTagMu.RUnlock()
	}
	if tag == nil {
		// Java throws if tagger missing; fail-closed drop match (do not invent POS).
		return nil
	}
	postagsListStr, ok := arguments["PostagsList"]
	if !ok {
		panic("Missing key 'PostagsList'")
	}
	filtered := (&CheckPostagsInSuggestionFilter{TagToken: tag}).Filter(match.GetSuggestedReplacements(), postagsListStr)
	if len(filtered) == 0 {
		return nil
	}
	match.SetSuggestedReplacements(filtered)
	return match
}

// Filter keeps multi-token suggestions whose tokens match postagsList (comma-separated regexes).
// Returns nil when none match (caller should suppress the rule match).
func (f *CheckPostagsInSuggestionFilter) Filter(replacements []string, postagsListStr string) []string {
	if f.TagToken == nil {
		return nil
	}
	postagsList := strings.Split(postagsListStr, ",")
	if len(postagsList) == 0 || (len(postagsList) == 1 && postagsList[0] == "") {
		return nil
	}
	var res []*regexp.Regexp
	for _, p := range postagsList {
		re, err := regexp.Compile(strings.TrimSpace(p))
		if err != nil {
			return nil
		}
		res = append(res, re)
	}
	var out []string
	for _, replacement := range replacements {
		tokens := strings.Fields(replacement)
		if len(tokens) != len(res) {
			// Java throws on mismatch; skip (fail-closed for that replacement).
			continue
		}
		ok := true
		for i, tok := range tokens {
			tags := f.TagToken(tok)
			match := false
			for _, tag := range tags {
				// Java: matchesPosTagRegex (substring/full per ATR); use MatchString on tags.
				if res[i].MatchString(tag) {
					match = true
					break
				}
			}
			if !match {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, replacement)
		}
	}
	return out
}
