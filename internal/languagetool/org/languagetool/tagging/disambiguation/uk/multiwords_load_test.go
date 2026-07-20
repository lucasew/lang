package uk

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestLoadUkrainianMultiwords(t *testing.T) {
	lines := LoadUkrainianMultiwordsLines()
	require.NotEmpty(t, lines)
	// all normalized to tab form
	for _, ln := range lines {
		require.Contains(t, ln, "\t", ln)
		parts := strings.SplitN(ln, "\t", 2)
		require.Len(t, parts, 2)
		require.NotEmpty(t, parts[0])
		require.NotEmpty(t, parts[1])
	}
	// known phrases
	joined := strings.Join(lines, "\n")
	require.Contains(t, joined, "а капела\tadv")
	require.Contains(t, joined, "на жаль\tinsert")
	require.Contains(t, joined, "як правило\tinsert")
}

func TestNormalizeUkrainianMultiwordLine(t *testing.T) {
	s, ok := normalizeUkrainianMultiwordLine("а капелаadv")
	require.True(t, ok)
	require.Equal(t, "а капела\tadv", s)
	s, ok = normalizeUkrainianMultiwordLine("без сумнівуinsert")
	require.True(t, ok)
	require.Equal(t, "без сумніву\tinsert", s)
	s, ok = normalizeUkrainianMultiwordLine("New York\tB-NP")
	require.True(t, ok)
	require.Equal(t, "New York\tB-NP", s)
	_, ok = normalizeUkrainianMultiwordLine("no-tag-here")
	require.False(t, ok)
}

func TestDefaultChunker_DisambiguatePhrase(t *testing.T) {
	c := NewDefaultUkrainianMultiwordChunker()
	require.NotNil(t, c)
	// "на жаль" → <insert> on both tokens (Java MultiWordChunker2 wrap)
	sent := atrSent2("на", "жаль")
	out := c.Disambiguate(sent)
	require.NotNil(t, out)
	toks := out.GetTokensWithoutWhitespace()
	require.GreaterOrEqual(t, len(toks), 3) // SENT_START + 2
	require.True(t, hasPOSSub(toks[1], "<insert>"), readingsDump(toks[1]))
	require.True(t, hasPOSSub(toks[2], "<insert>"), readingsDump(toks[2]))
}

func hasPOSSub(atr *languagetool.AnalyzedTokenReadings, sub string) bool {
	if atr == nil {
		return false
	}
	for _, r := range atr.GetReadings() {
		if r != nil && r.GetPOSTag() != nil && strings.Contains(*r.GetPOSTag(), sub) {
			return true
		}
	}
	return false
}

func readingsDump(atr *languagetool.AnalyzedTokenReadings) string {
	if atr == nil {
		return ""
	}
	var b []string
	for _, r := range atr.GetReadings() {
		if r != nil && r.GetPOSTag() != nil {
			b = append(b, *r.GetPOSTag())
		}
	}
	return strings.Join(b, ",")
}

func TestNewUkrainianHybrid_LoadsMultiwords(t *testing.T) {
	d := NewUkrainianHybridDisambiguator()
	require.NotNil(t, d.Chunker)
	// default hybrid should not use empty chunker
	lines := LoadUkrainianMultiwordsLines()
	require.NotEmpty(t, lines)
}

func atrSent2(a, b string) *languagetool.AnalyzedSentence {
	// SENT_START + two tokens
	pos := "prep"
	pos2 := "noun:inanim:n:v_naz"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(a, &pos, &a)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(b, &pos2, &b)),
	}
	// fix start token
	st := languagetool.SentenceStartTagName
	toks[0] = languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &st, nil))
	return languagetool.NewAnalyzedSentence(toks)
}
