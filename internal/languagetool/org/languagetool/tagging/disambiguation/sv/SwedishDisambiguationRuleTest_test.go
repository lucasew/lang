package sv

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	"github.com/stretchr/testify/require"
)

func TestSwedishDisambiguationRule_Chunker(t *testing.T) {
	d := NewSwedishHybridDisambiguator()
	d.Inner = disambiguation.NewMultiWordChunker([]string{"Stock holm\tNP"}, disambiguation.MultiWordChunkerSettings{})
	// identity when no match
	s := languagetool.AnalyzePlain("Hej världen")
	out := d.Disambiguate(s)
	require.NotNil(t, out)
	require.Equal(t, s.GetText(), out.GetText())
}
