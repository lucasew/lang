package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	"github.com/stretchr/testify/require"
)

func TestFrenchRuleDisambiguator_Chunker(t *testing.T) {
	c := disambiguation.NewMultiWordChunker([]string{"New York\tB-NP"}, disambiguation.MultiWordChunkerSettings{AllowFirstCapitalized: true})
	s := languagetool.AnalyzePlain("Bonjour le monde")
	out := c.Disambiguate(s)
	require.NotNil(t, out)
}
