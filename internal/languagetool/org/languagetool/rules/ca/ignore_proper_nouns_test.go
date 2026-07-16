package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestIgnoreProperNouns(t *testing.T) {
	r := NewIgnoreProperNouns()
	require.Equal(t, "IGNORE_PROPER_NOUNS", r.GetID())
	require.Equal(t, 0, r.MinToCheckParagraph())

	np := "NPMS000"
	s1 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Barcelona", &np, strPtr("Barcelona"))),
	})
	s1.GetTokens()[1].SetStartPos(0)

	s2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Barcelona", nil, nil)),
	})
	s2.GetTokens()[1].SetStartPos(0)

	matches := r.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.Len(t, matches, 1)
	// offset includes s1 corrected length ("" + "Barcelona" = 9)
	require.Equal(t, s1.GetCorrectedTextLength(), matches[0].GetFromPos())

	require.Empty(t, r.MatchList([]*languagetool.AnalyzedSentence{s1}))
}

func strPtr(s string) *string { return &s }
