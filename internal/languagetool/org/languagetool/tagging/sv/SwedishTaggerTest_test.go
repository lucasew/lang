package sv

// Twin of SwedishTaggerTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestSwedishTagger_Dictionary(t *testing.T) {
	wt := tagging.MapWordTagger{"hus": {tagging.NewTaggedWord("hus", "NN")}}
	tagger := NewSwedishTagger(wt)
	require.Equal(t, SwedishDictPath, tagger.GetDictionaryPath())
	require.Len(t, tagger.TagWord("hus"), 1)
}

func TestSwedishTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"detta": {tagging.NewTaggedWord("detta", "PN")},
		"test":  {tagging.NewTaggedWord("test", "NN")},
	}
	got := NewSwedishTagger(wt).Tag([]string{"Detta", "test", "xyz"})
	require.Len(t, got, 3)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Nil(t, got[2].GetReadings()[0].GetPOSTag())
}
