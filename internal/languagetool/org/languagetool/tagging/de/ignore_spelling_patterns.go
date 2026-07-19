package de

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Patterns for tokens Java MultitokenIgnore marks ignore-by-speller.
// Incomplete vs full German multitoken XML lists — only the documented
// digit-hyphen adjective pattern (GermanDisambiguationTest "3-adische"), not invent.
var (
	// digit-hyphen adjective stems: "3-adische", "2-fach"
	reDigitHyphenAdj = regexp.MustCompile(`(?i)^\d+-[a-zäöüß]+$`)
)

// MarkIgnoreSpellingPatterns marks tokens matching reDigitHyphenAdj.
// Incomplete MultitokenIgnore subset for GermanRuleDisambiguator tests.
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
