package nl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// PreDisambiguate is AbstractDisambiguator identity (Java DutchHybrid has no override).
// Full Disambiguate outcomes: dutch_hybrid_disambiguator_order_test.go.
func TestDutchHybridDisambiguator_PreDisambiguateIdentity(t *testing.T) {
	d := NewDutchHybridDisambiguator()
	s := languagetool.AnalyzePlain("Hallo wereld")
	require.Equal(t, s, d.PreDisambiguate(s))
}
