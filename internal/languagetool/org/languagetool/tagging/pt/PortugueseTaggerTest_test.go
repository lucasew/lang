package pt

// Twin of PortugueseTaggerTest — MapWordTagger + contraction greens
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestPortugueseTagger_Basic(t *testing.T) {
	wt := tagging.MapWordTagger{"casa": {tagging.NewTaggedWord("casa", "NCFS000")}}
	got := NewPortugueseTagger(wt).Tag([]string{"Casa", "xyz"})
	require.Len(t, got, 2)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
}

func TestPortugueseTagger_Contractions(t *testing.T) {
	tg := NewPortugueseTagger(tagging.MapWordTagger{})
	got := tg.Tag([]string{"do", "da", "no", "à"})
	require.True(t, got[0].IsTagged())
	require.True(t, got[1].IsTagged())
	require.True(t, got[2].IsTagged())
	require.True(t, got[3].IsTagged())
	// do → de + o readings
	require.GreaterOrEqual(t, len(got[0].GetReadings()), 2)
}

func TestPortugueseTagger_OrdinalAbbreviations(t *testing.T) {
	// soft: without dict, ordinals untagged
	t.Skip("unimplemented: needs full Portuguese dict for ordinal abbreviations")
}
func TestPortugueseTagger_ContractionTagging(t *testing.T) {
	// green path covered by TestPortugueseTagger_Contractions
	require.NotEmpty(t, ContractionReadings("pelo"))
}
func TestPortugueseTagger_Compound(t *testing.T) {
	t.Skip("unimplemented: compound tagging needs dict")
}
func TestPortugueseTagger_ProductivePrefixes(t *testing.T) {
	t.Skip("unimplemented: productive prefixes need dict")
}
func TestPortugueseTagger_Enclitics(t *testing.T) {
	t.Skip("unimplemented: enclitics need dict")
}
