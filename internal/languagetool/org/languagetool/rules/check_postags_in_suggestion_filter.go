package rules

import (
	"regexp"
	"strings"
)

// CheckPostagsInSuggestionFilter ports org.languagetool.rules.CheckPostagsInSuggestionFilter.
// TagToken returns POS tags for a single token; nil keeps no suggestions (suppress).
type CheckPostagsInSuggestionFilter struct {
	TagToken func(token string) []string
}

func NewCheckPostagsInSuggestionFilter(tag func(string) []string) *CheckPostagsInSuggestionFilter {
	return &CheckPostagsInSuggestionFilter{TagToken: tag}
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
			continue // Java throws; we skip malformed
		}
		ok := true
		for i, tok := range tokens {
			tags := f.TagToken(tok)
			match := false
			for _, tag := range tags {
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
