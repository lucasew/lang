package multitoken

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// MultitokenSpellerFilter ports
// org.languagetool.rules.spelling.multitoken.MultitokenSpellerFilter.
// Speller and misspelled hooks are pluggable (no full Language stack).
type MultitokenSpellerFilter struct {
	Speller    *MultitokenSpeller
	// IsMisspelled optional: when set and returns true for the error string,
	// GetSuggestions is called with areTokensAcceptedBySpeller=false.
	IsMisspelled func(s string) bool
	// AtSentenceStart when true capitalizes lower-case suggestions.
	AtSentenceStart bool
}

// AcceptRuleMatch attaches multiword suggestions to match; returns nil to drop.
func (f *MultitokenSpellerFilter) AcceptRuleMatch(match *rules.RuleMatch, originalError string) *rules.RuleMatch {
	if f == nil || match == nil || f.Speller == nil {
		return match
	}
	if originalError == "" && match.Sentence != nil {
		// fall back to covered text if available via positions
		text := match.Sentence.GetText()
		if match.FromPos >= 0 && match.ToPos <= len(text) && match.FromPos < match.ToPos {
			originalError = text[match.FromPos:match.ToPos]
		}
	}
	if originalError == "" {
		return nil
	}
	acceptedBySpeller := false
	if f.IsMisspelled != nil {
		acceptedBySpeller = !f.IsMisspelled(originalError)
	}
	replacements := f.Speller.GetSuggestionsOpts(originalError, acceptedBySpeller)
	if len(replacements) == 0 {
		return nil
	}
	if len(originalError) > 4 && isAllUpper(originalError) {
		up := make([]string, 0, len(replacements))
		seen := map[string]struct{}{}
		for _, r := range replacements {
			n := strings.ToUpper(r)
			if n == originalError {
				continue
			}
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			up = append(up, n)
		}
		replacements = up
	} else if f.AtSentenceStart {
		cap := make([]string, 0, len(replacements))
		seen := map[string]struct{}{}
		for _, r := range replacements {
			n := r
			if r == strings.ToLower(r) {
				n = uppercaseFirst(r)
			}
			if n == originalError {
				continue
			}
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			cap = append(cap, n)
		}
		replacements = cap
	}
	if len(replacements) == 0 {
		return nil
	}
	match.SetSuggestedReplacements(replacements)
	return match
}

func isAllUpper(s string) bool {
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

func uppercaseFirst(s string) string {
	rs := []rune(s)
	if len(rs) == 0 {
		return s
	}
	rs[0] = unicode.ToUpper(rs[0])
	return string(rs)
}
