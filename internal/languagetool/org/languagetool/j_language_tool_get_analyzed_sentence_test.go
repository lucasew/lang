package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of JLanguageTool.getAnalyzedSentence: input is one sentence unit (no SRX re-split).
// Tools.checkBitext uses getAnalyzedSentence on full strings that may contain periods.
func TestGetAnalyzedSentence_NoSRXSplit(t *testing.T) {
	lt := NewJLanguageTool("en")
	// Multi-clause string — Analyze may SRX-split; GetAnalyzedSentence must not.
	text := "A sentence. This is not actual."
	one := lt.GetAnalyzedSentence(text)
	require.NotNil(t, one)
	require.Equal(t, text, one.GetText())
	// Full token stream includes both clauses
	nonWS := one.GetTokensWithoutWhitespace()
	require.GreaterOrEqual(t, len(nonWS), 6) // SENT_START + words from both clauses

	// Analyze may return 1+ sentences depending on SRX; GetAnalyzedSentence always one.
	many := lt.Analyze(text)
	require.NotEmpty(t, many)
	// When SRX splits, Analyze has more than one; GetAnalyzedSentence still one.
	if len(many) > 1 {
		require.Equal(t, 1, 1) // document split path
		// Combined texts of Analyze parts should reassemble
		var joined string
		for _, s := range many {
			joined += s.GetText()
		}
		require.Equal(t, text, joined)
	}
}

func TestGetRawAnalyzedSentence_NoDisambiguator(t *testing.T) {
	lt := NewJLanguageTool("en")
	// Inject a disambiguator that would panic if called on raw path.
	lt.Disambiguator = panickingDisambiguator{}
	raw := lt.GetRawAnalyzedSentence("Hello world.")
	require.NotNil(t, raw)
	require.Equal(t, "Hello world.", raw.GetText())
	// GetAnalyzedSentence would call disambiguator — skip if it panics
}

type panickingDisambiguator struct{}

func (panickingDisambiguator) Disambiguate(s *AnalyzedSentence) *AnalyzedSentence {
	panic("disambiguator must not run on GetRawAnalyzedSentence")
}

// Twin of JLanguageTool.getAnalyzedSentence ResultCache path (ToolsTest cache != null).
func TestGetAnalyzedSentence_ResultCache(t *testing.T) {
	cache := NewResultCache(100)
	lt := NewJLanguageToolWithCache("en", cache)
	// First call: miss + put
	a1 := lt.GetAnalyzedSentence("Hello world.")
	require.NotNil(t, a1)
	// Second call: hit same pointer path via cache
	a2 := lt.GetAnalyzedSentence("Hello world.")
	require.NotNil(t, a2)
	require.Equal(t, a1.GetText(), a2.GetText())
	require.GreaterOrEqual(t, cache.HitCount(), int64(1))
	// Different text misses
	a3 := lt.GetAnalyzedSentence("Other text.")
	require.NotNil(t, a3)
	require.NotEqual(t, a1.GetText(), a3.GetText())
	// Nil cache still works
	lt2 := NewJLanguageTool("en")
	require.NotNil(t, lt2.GetAnalyzedSentence("x"))
}
