package rules

import (
	"fmt"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// immunizeAllThe immunizes "the" on GetTokens() after Java-style copy.
// Real DisambiguationPatternRule.Replace rebuilds nonBlank via NewAnalyzedSentenceFull.
type immunizeAllThe struct{}

func (immunizeAllThe) Replace(s *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if s == nil {
		return nil
	}
	// Rebuild like Java DisambiguationPatternRule.Replace return path:
	// mutate full tokens then NewAnalyzedSentence so nonBlank is recomputed.
	toks := append([]*languagetool.AnalyzedTokenReadings(nil), s.GetTokens()...)
	for _, t := range toks {
		if t != nil && t.GetToken() == "the" {
			t.Immunize(0)
		}
	}
	return languagetool.NewAnalyzedSentence(toks)
}

func TestSentenceWithImmunization_EmptyIdentity(t *testing.T) {
	s := languagetool.AnalyzePlain("the the")
	out := SentenceWithImmunization(s, nil)
	require.Same(t, s, out)
}

func TestWordRepeatRule_ImmunizedNoMatch(t *testing.T) {
	s := languagetool.AnalyzePlain("the the")
	// baseline without anti-patterns
	r2 := NewWordRepeatRule(map[string]string{"repetition": "repeat"})
	ms2 := r2.Match(s)
	for _, tok := range s.GetTokensWithoutWhitespace() {
		fmt.Printf("baseline tok %q imm=%v\n", tok.GetToken(), tok.IsImmunized())
	}
	require.NotEmpty(t, ms2, "baseline double 'the' must match WordRepeatRule")

	r := NewWordRepeatRule(map[string]string{"repetition": "repeat"})
	r.AntiPatterns = []SentenceReplacer{immunizeAllThe{}}
	ms := r.Match(s)
	require.Empty(t, ms, "immunized word repeat must not fire")
}
