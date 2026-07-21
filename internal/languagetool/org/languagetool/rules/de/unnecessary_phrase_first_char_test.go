package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Twin of UnnecessaryPhraseRule.firstCharToLower — UTF-16 substring(0,1).toLowerCase().
func TestFirstCharToLowerPhrase_UTF16(t *testing.T) {
	// sentence start token index 1 is the first content word
	ss := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil), 0)
	// "Über" must become "über" not corrupt first byte of UTF-8
	ueber := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("Über", nil, nil), 0)
	toks := []*languagetool.AnalyzedTokenReadings{ss, ueber}
	require.Equal(t, "über", firstCharToLowerPhrase(toks, 1))

	// nToken != 1 → surface unchanged (even if capitalized umlaut)
	mid := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("x", nil, nil), 0)
	toks2 := []*languagetool.AnalyzedTokenReadings{ss, mid, ueber}
	require.Equal(t, "Über", firstCharToLowerPhrase(toks2, 2))
	// length < 2 UTF-16: single letter unchanged
	a := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("A", nil, nil), 0)
	require.Equal(t, "A", firstCharToLowerPhrase([]*languagetool.AnalyzedTokenReadings{ss, a}, 1))
	// multi-letter ASCII
	in := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("In", nil, nil), 0)
	require.Equal(t, "in", firstCharToLowerPhrase([]*languagetool.AnalyzedTokenReadings{ss, in}, 1))
}
