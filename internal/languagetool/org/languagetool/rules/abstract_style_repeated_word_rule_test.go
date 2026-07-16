package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAbstractStyleRepeatedWordRule(t *testing.T) {
	r := NewAbstractStyleRepeatedWordRule()
	s1 := languagetool.AnalyzePlain("The cat sat near the cat.")
	matches := r.MatchList([]*languagetool.AnalyzedSentence{s1})
	require.NotEmpty(t, matches)

	s2a := languagetool.AnalyzePlain("Dogs are nice.")
	s2b := languagetool.AnalyzePlain("Dogs run fast.")
	m2 := r.MatchList([]*languagetool.AnalyzedSentence{s2a, s2b})
	require.NotEmpty(t, m2)
}
