package pt

// Twin of languagetool-language-modules/pt PortugueseTagger tests (MapWordTagger smokes).

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
	require.Equal(t, "NCFS000", *got[0].GetReadings()[0].GetPOSTag())
	// unknown → null POS
	require.Nil(t, got[1].GetReadings()[0].GetPOSTag())
}

func TestPortugueseTagger_Ordinals(t *testing.T) {
	tg := NewPortugueseTagger(tagging.MapWordTagger{})
	got := tg.Tag([]string{"1.º", "2.ª"})
	require.True(t, got[0].IsTagged())
	require.True(t, got[1].IsTagged())
	// noun + adj readings
	require.GreaterOrEqual(t, len(got[0].GetReadings()), 2)
}

func TestPortugueseTagger_PercentDegree(t *testing.T) {
	tg := NewPortugueseTagger(tagging.MapWordTagger{})
	got := tg.Tag([]string{"10%", "20°"})
	require.Equal(t, "NCMP000", *got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "NCMP000", *got[1].GetReadings()[0].GetPOSTag())
}

func TestPortugueseTagger_Mente(t *testing.T) {
	// possibleAdj = lowerWord without "mente"; ADJ FS tag → RG
	wt := tagging.MapWordTagger{
		"rapida": {tagging.NewTaggedWord("rapido", "AQ0FS0")},
	}
	got := NewPortugueseTagger(wt).Tag([]string{"rapidamente"})
	require.True(t, got[0].IsTagged())
	require.Equal(t, "RG", *got[0].GetReadings()[0].GetPOSTag())
}

func TestPortugueseTagger_SotoPrefix(t *testing.T) {
	wt := tagging.MapWordTagger{
		"por": {tagging.NewTaggedWord("por", "VMIP1S0")},
	}
	got := NewPortugueseTagger(wt).Tag([]string{"soto-por"})
	require.True(t, got[0].IsTagged())
	require.Equal(t, "VMIP1S0", *got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "soto-por", *got[0].GetReadings()[0].GetLemma())
}

func TestPortugueseTagger_DictionaryPath(t *testing.T) {
	require.Equal(t, PortugueseDictPath, NewPortugueseTagger(nil).GetDictionaryPath())
}
