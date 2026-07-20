package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestExtractSuggestions_XMLOnly(t *testing.T) {
	clean, sugs := extractSuggestions(`Use <suggestion>foo</suggestion> not bar`)
	require.Equal(t, "Use foo not bar", clean)
	require.Equal(t, []string{"foo"}, sugs)
	// Quoted prose is not invented as a suggestion.
	clean2, sugs2 := extractSuggestions(`Did you mean "hello"?`)
	require.Equal(t, `Did you mean "hello"?`, clean2)
	require.Empty(t, sugs2)
}

func TestExpandPatternBackrefs(t *testing.T) {
	require.Equal(t, "hello world", expandPatternBackrefs(`\1 \2`, []string{"hello", "world"}))
	require.Equal(t, `\9`, expandPatternBackrefs(`\9`, []string{"a"})) // out of range stays
	require.Equal(t, "plain", expandPatternBackrefs("plain", nil))
}

func TestMatchSpanTokens_NoSentStartInvent(t *testing.T) {
	start := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &start, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("could", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("of", nil, nil), 6),
	}
	require.True(t, toks[0].IsSentenceStart())
	sent := testSentence(toks...)
	span := matchSpanTokens(sent, 0, 8)
	// Only real surfaces — no empty invent for SENT_START
	require.Equal(t, []string{"could", "of"}, span)
	require.Equal(t, "could of", expandPatternBackrefs(`\1 \2`, span))
}
