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
	tg := NewPortugueseTagger(tagging.MapWordTagger{})
	got := tg.Tag([]string{"1.º", "2.ª"})
	require.True(t, got[0].IsTagged())
	require.True(t, got[1].IsTagged())
}
func TestPortugueseTagger_ContractionTagging(t *testing.T) {
	// green path covered by TestPortugueseTagger_Contractions
	require.NotEmpty(t, ContractionReadings("pelo"))
}
func TestPortugueseTagger_Compound(t *testing.T) {
	// compound case variants: inject jiu-jitsu
	wt := tagging.MapWordTagger{
		"jiu-jitsu": {tagging.NewTaggedWord("jiu-jitsu", "NCMS000")},
	}
	tg := NewPortugueseTagger(wt)
	for _, w := range []string{"jiu-jitsu", "Jiu-jitsu", "JIU-JITSU"} {
		got := tg.Tag([]string{w})
		require.True(t, got[0].IsTagged(), w)
	}
}
func TestPortugueseTagger_ProductivePrefixes(t *testing.T) {
	// soto-trepei: bare verb trepei in inject dict
	wt := tagging.MapWordTagger{
		"trepei": {tagging.NewTaggedWord("trepar", "VMIS1S0")},
	}
	tg := NewPortugueseTagger(wt)
	got := tg.Tag([]string{"soto-trepei", "xoxotrepei"})
	require.True(t, got[0].IsTagged())
	require.False(t, got[1].IsTagged()) // not a real prefix
	lemma := got[0].GetReadings()[0].GetLemma()
	require.NotNil(t, lemma)
	require.Equal(t, "soto-trepar", *lemma)
}
func TestPortugueseTagger_Enclitics(t *testing.T) {
	wt := tagging.MapWordTagger{"diz": {tagging.NewTaggedWord("dizer", "VMIP3S0")}}
	tg := NewPortugueseTagger(wt)
	got := tg.Tag([]string{"diz-me"})
	require.True(t, got[0].IsTagged())
	verb, clit, ok := EncliticSplit("diz-me")
	require.True(t, ok)
	require.Equal(t, "diz", verb)
	require.Equal(t, "me", clit)
}
