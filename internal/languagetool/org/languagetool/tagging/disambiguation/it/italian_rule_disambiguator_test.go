package it

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestItalianRuleDisambiguator(t *testing.T) {
	d := NewItalianRuleDisambiguator()
	s := languagetool.AnalyzePlain("Ciao mondo")
	require.Equal(t, s, d.Disambiguate(s))
}
