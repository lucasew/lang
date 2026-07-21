package de

// Twin of StyleRepeatedVeryShortSentences (Java MIN_WORDS=4, MIN_REPEATED=3, example pair).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestStyleRepeatedVeryShortSentences(t *testing.T) {
	rule := NewStyleRepeatedVeryShortSentences(nil)
	require.Equal(t, "STYLE_REPEATED_SHORT_SENTENCES", rule.GetID())
	require.Equal(t, "Stakkato-Sätze", rule.GetDescription())
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, 4, rule.MinWords)
	require.Equal(t, 3, rule.MinRepeated)
	require.True(t, rule.ExcludeDirectSpeech)
	require.Equal(t, 3, rule.MinToCheckParagraph())
	require.NotEmpty(t, rule.GetIncorrectExamples())

	// Java example: three short sentences
	// tokens.length > 3 && <= minWords+2 (SENT_START + words + punct)
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Das Auto kam näher."),
		languagetool.AnalyzePlain("Der Hund schlief."),
		languagetool.AnalyzePlain("Die Reifen quietschten."),
	}
	for _, s := range sents {
		n := len(s.GetTokensWithoutWhitespace())
		require.Greater(t, n, 3, s.GetText())
		require.LessOrEqual(t, n, rule.MinWords+2, s.GetText())
	}
	ms := rule.MatchList(sents)
	require.Equal(t, 3, len(ms))
	for _, m := range ms {
		require.Equal(t, "Stakkato-Sätze", m.GetMessage())
		require.Greater(t, m.GetToPos(), m.GetFromPos())
	}
	// Java: start = tokens[len-2].startPos (+pos), end = tokens[len-1].endPos (+pos)
	// first sentence has pos=0
	toks0 := sents[0].GetTokensWithoutWhitespace()
	require.Equal(t, toks0[len(toks0)-2].GetStartPos(), ms[0].GetFromPos())
	require.Equal(t, toks0[len(toks0)-1].GetEndPos(), ms[0].GetToPos())

	// long sentence breaks the streak
	sents2 := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Er lief."),
		languagetool.AnalyzePlain("Sie rief."),
		languagetool.AnalyzePlain("Dieser deutlich längere Satz unterbricht die Serie von kurzen Sätzen hier."),
		languagetool.AnalyzePlain("Er ging."),
	}
	require.Equal(t, 0, len(rule.MatchList(sents2)))

	// fewer than minRepeated → 0
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Er lief."),
		languagetool.AnalyzePlain("Sie rief."),
	})))

	// direct speech exclusion
	quoted := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("„Das Auto kam näher.“"),
		languagetool.AnalyzePlain("„Der Hund schlief.“"),
		languagetool.AnalyzePlain("„Die Reifen quietschten.“"),
	}
	require.Equal(t, 0, len(rule.MatchList(quoted)), "quoted short sentences excluded")
}
