package nl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDutchHybridDisambiguator(t *testing.T) {
	d := NewDutchHybridDisambiguator()
	s := languagetool.AnalyzePlain("Hallo wereld")
	require.Equal(t, s, d.Disambiguate(s))
	require.Equal(t, s, d.PreDisambiguate(s))
}
