package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseDisambiguationRule_Chunker(t *testing.T) {
	d := NewPortugueseHybridDisambiguator()
	s := languagetool.AnalyzePlain("Olá mundo")
	out := d.Disambiguate(s)
	require.NotNil(t, out)
}
