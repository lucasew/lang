package eo

// Twin of languagetool-language-modules/eo/src/test/java/org/languagetool/tagging/eo/EsperantoTaggerTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

// Port of EsperantoTaggerTest.testTagger
func TestEsperantoTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"domo":  {tagging.NewTaggedWord("domo", "S")},
		"granda": {tagging.NewTaggedWord("granda", "A")},
	}
	got := NewEsperantoTagger(wt).Tag([]string{"Domo", "granda", "xyz"})
	require.Len(t, got, 3)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.NotNil(t, got[1].GetReadings()[0].GetPOSTag())
	require.Nil(t, got[2].GetReadings()[0].GetPOSTag())
}
