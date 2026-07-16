package ro

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	"github.com/stretchr/testify/require"
)

func disambigSmoke(t *testing.T, text string) {
	t.Helper()
	c := disambiguation.NewMultiWordChunker(nil, disambiguation.MultiWordChunkerSettings{})
	out := c.Disambiguate(languagetool.AnalyzePlain(text))
	require.NotNil(t, out)
	require.Equal(t, languagetool.AnalyzePlain(text).GetText(), out.GetText())
}

func TestRomanianRuleDisambiguator_Care1(t *testing.T) {
	disambigSmoke(t, "Care este asta?")
}

func TestRomanianRuleDisambiguator_EsteO(t *testing.T) {
	disambigSmoke(t, "Este o carte.")
}

func TestRomanianRuleDisambiguator_DezambiguizareVerb(t *testing.T) {
	disambigSmoke(t, "Eu merg acasă.")
}
