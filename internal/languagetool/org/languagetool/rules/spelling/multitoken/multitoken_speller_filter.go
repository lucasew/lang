package multitoken

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// MultitokenSpellerFilter ports
// org.languagetool.rules.spelling.multitoken.MultitokenSpellerFilter.
// Speller and misspelled hooks are pluggable (no full Language stack).
type MultitokenSpellerFilter struct {
	Speller *MultitokenSpeller
	// IsMisspelled optional token-level speller (Java SpellingCheckRule.isMisspelled).
	// When set, isMisspelled tokenizes the error with WordTokenizer and ORs token results
	// (Java MultitokenSpellerFilter.isMisspelled). Nil → null SpellingCheckRule path.
	IsMisspelled func(token string) bool
	// Tokenize ports language.getWordTokenizer().tokenize; nil → WordTokenizer.
	Tokenize func(s string) []string
	// AtSentenceStart when true capitalizes lower-case suggestions.
	AtSentenceStart bool
	// CheckSpelling enables Java en/de/pt/nl areTokensAcceptedBySpeller path.
	// When false (default), areTokensAcceptedBySpeller stays false like non-en/de/pt/nl.
	// When true and IsMisspelled is nil, acceptedBySpeller is true (null speller → !false).
	// When true and IsMisspelled is set, acceptedBySpeller = !isMisspelled(error).
	// If IsMisspelled is set without CheckSpelling, CheckSpelling is treated as true
	// (host wired a speller for the en/de/pt/nl path).
	CheckSpelling bool
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
	// Java: areTokensAcceptedBySpeller = false unless shortCode in {en,de,pt,nl}
	// then = !isMisspelled(underlinedError, lang).
	acceptedBySpeller := false
	checkSpell := f.CheckSpelling || f.IsMisspelled != nil
	if checkSpell {
		// null SpellingCheckRule → isMisspelled false → accepted true
		acceptedBySpeller = !f.isMisspelled(originalError)
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

// isMisspelled ports MultitokenSpellerFilter.isMisspelled(String, Language):
//
//	if spellerRule == null → false
//	tokens = wordTokenizer.tokenize(s)
//	any token misspelled → true
func (f *MultitokenSpellerFilter) isMisspelled(s string) bool {
	if f == nil || f.IsMisspelled == nil {
		// Java: null SpellingCheckRule → false (not misspelled)
		return false
	}
	tokens := f.tokenizeError(s)
	if len(tokens) == 0 {
		return false
	}
	for _, tok := range tokens {
		if f.IsMisspelled(tok) {
			return true
		}
	}
	return false
}

func (f *MultitokenSpellerFilter) tokenizeError(s string) []string {
	if f != nil && f.Tokenize != nil {
		return f.Tokenize(s)
	}
	return tokenizers.NewWordTokenizer().Tokenize(s)
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
