package sr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSerbianHybridDisambiguator(t *testing.T) {
	d := NewSerbianHybridDisambiguator()
	s := languagetool.AnalyzePlain("Zdravo svete")
	require.Equal(t, s, d.Disambiguate(s))
}
