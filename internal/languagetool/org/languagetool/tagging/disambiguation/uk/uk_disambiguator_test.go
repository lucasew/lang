package uk

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUkrainianHybridDisambiguator(t *testing.T) {
	d := NewUkrainianHybridDisambiguator()
	require.Nil(t, d.Disambiguate(nil))
	s := languagetool.NewAnalyzedSentence(nil)
	require.NotNil(t, d.Disambiguate(s))
	require.NotNil(t, NewSimpleDisambiguator().Disambiguate(s))
	d2 := NewUkrainianHybridDisambiguatorWith(NewUkrainianMultiwordChunker(nil), NewSimpleDisambiguator())
	require.NotNil(t, d2.Disambiguate(s))
}

func TestUkrainianMultiwordChunker_POSRegexMatch(t *testing.T) {
	// phrase with /POS-regex second token (Java UkrainianMultiwordChunker.matches)
	ch := NewUkrainianMultiwordChunker([]string{"з /noun:.*v_oru.*\tprep_oru"})
	require.NotNil(t, ch)
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil),
	}, 0)
	p1, l1 := "prep", "з"
	z := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("з", &p1, &l1),
	}, 0)
	p2, l2 := "noun:inanim:m:v_oru", "домом"
	n := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("домом", &p2, &l2),
	}, 0)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, z, n})
	out := ch.Disambiguate(sent)
	require.NotNil(t, out)
	// first content token should gain multiword tag
	toks := out.GetTokensWithoutWhitespace()
	require.GreaterOrEqual(t, len(toks), 2)
	// WrapTag default wraps as <prep_oru>
	found := false
	for _, t2 := range toks[1].GetReadings() {
		if t2 != nil && t2.GetPOSTag() != nil && strings.Contains(*t2.GetPOSTag(), "prep_oru") {
			found = true
		}
	}
	require.True(t, found, "expected multiword prep_oru tag after POS-regex match")
}

func TestUkrainianMultiwordChunker_POSRegexNoMatch(t *testing.T) {
	// POS does not match /noun:.*v_oru.* → no multiword wrap
	ch := NewUkrainianMultiwordChunker([]string{"з /noun:.*v_oru.*\tprep_oru"})
	require.NotNil(t, ch)
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil),
	}, 0)
	p1, l1 := "prep", "з"
	z := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("з", &p1, &l1),
	}, 0)
	p2, l2 := "noun:inanim:m:v_naz", "дім"
	n := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("дім", &p2, &l2),
	}, 0)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, z, n})
	out := ch.Disambiguate(sent)
	require.NotNil(t, out)
	toks := out.GetTokensWithoutWhitespace()
	require.GreaterOrEqual(t, len(toks), 2)
	for _, t2 := range toks[1].GetReadings() {
		if t2 != nil && t2.GetPOSTag() != nil {
			require.NotContains(t, *t2.GetPOSTag(), "prep_oru")
		}
	}
}
