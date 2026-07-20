package sv

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	"github.com/stretchr/testify/require"
)

// Twin of SwedishDisambiguationRuleTest.testChunker setup: MultiWordChunker on hybrid.
// Java SwedishHybridDisambiguator: chunker.disambiguate(disambiguator.disambiguate(input)).
func TestSwedishDisambiguationRule_Chunker(t *testing.T) {
	d := NewSwedishHybridDisambiguator()
	// Java field "chunker" = MultiWordChunker; Go exposes Chunker (not Inner).
	d.Chunker = disambiguation.NewMultiWordChunker([]string{"Stock holm\tNP"}, disambiguation.MultiWordChunkerSettings{})
	// identity when no multiword match
	s := languagetool.AnalyzePlain("Hej världen")
	out := d.Disambiguate(s)
	require.NotNil(t, out)
	require.Equal(t, s.GetText(), out.GetText())
}
