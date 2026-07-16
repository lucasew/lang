package ga

// Twin of IrishTaggerTest — full GA dict deferred; MapWordTagger smoke.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestIrishTagger_NoTests(t *testing.T) {
	// Java class has no @Test methods; exercise surface anyway.
	wt := tagging.MapWordTagger{
		"madra": {tagging.NewTaggedWord("madra", "Noun:Masc:Com:Sg")},
		"tá":    {tagging.NewTaggedWord("bí", "Verb:PresInd")},
	}
	tagger := NewIrishTagger(wt)
	require.Equal(t, IrishDictPath, tagger.GetDictionaryPath())
	got := tagger.Tag([]string{"Tá", "madra", "xyz"})
	require.Len(t, got, 3)
	// lowercasing lookup for "Tá"
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "Noun:Masc:Com:Sg", *got[1].GetReadings()[0].GetPOSTag())
	require.Nil(t, got[2].GetReadings()[0].GetPOSTag())
}
