package ga

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	"github.com/stretchr/testify/require"
)

func TestIrishDisambiguationRule_Disambiguation(t *testing.T) {
	c := disambiguation.NewMultiWordChunker([]string{"New York\tNP"}, disambiguation.MultiWordChunkerSettings{AllowFirstCapitalized: true})
	out := c.Disambiguate(languagetool.AnalyzePlain("New York is big"))
	require.NotNil(t, out)
}
