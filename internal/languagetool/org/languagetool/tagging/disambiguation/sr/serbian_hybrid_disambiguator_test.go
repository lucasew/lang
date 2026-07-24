package sr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Smoke: hybrid constructs and is a no-op on plain untagged non-Roman input.
// Stage-order outcome twins live in serbian_hybrid_disambiguator_order_test.go.
func TestSerbianHybridDisambiguator(t *testing.T) {
	d := NewSerbianHybridDisambiguator()
	s := languagetool.AnalyzePlain("Zdravo svete")
	require.Equal(t, s, d.Disambiguate(s))
}

func TestSerbianHybridDisambiguator_ConstructorWiresWhenResourcesPresent(t *testing.T) {
	if DiscoverSerbianMultiwords() == "" || DiscoverSerbianDisambiguationXML() == "" {
		t.Skip("official SR hybrid resources not discoverable")
	}
	d := NewSerbianHybridDisambiguator()
	require.NotNil(t, d.Chunker, "Java final field chunker")
	require.NotNil(t, d.Rules, "Java final field disambiguator")
}
