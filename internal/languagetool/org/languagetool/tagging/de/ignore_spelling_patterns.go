package de

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Patterns for tokens often marked ignore-by-speller via multiword/XML rules
// (simplified stand-in for full German multitoken lists).
var (
	// digit-hyphen adjective stems: "3-adische", "2-fach"
	reDigitHyphenAdj = regexp.MustCompile(`(?i)^\d+-[a-zäöüß]+$`)
)

// MarkIgnoreSpellingPatterns marks tokens matching known German ignore patterns.
// Used as a MultitokenIgnore stage stand-in for GermanRuleDisambiguator.
func MarkIgnoreSpellingPatterns(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	toks := input.GetTokens()
	// work in place on tokens
	for _, t := range toks {
		if t == nil || t.IsWhitespace() {
			continue
		}
		if reDigitHyphenAdj.MatchString(t.GetToken()) {
			t.IgnoreSpelling()
		}
	}
	return input
}

// IsDigitHyphenAdj reports the pattern used for 3-adische style tokens.
func IsDigitHyphenAdj(token string) bool {
	return reDigitHyphenAdj.MatchString(token)
}

// ignoreSpellingStep adapts MarkIgnoreSpellingPatterns for GermanRuleDisambiguator tests.
type ignoreSpellingStep struct{}

func (ignoreSpellingStep) Disambiguate(s *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return MarkIgnoreSpellingPatterns(s)
}
