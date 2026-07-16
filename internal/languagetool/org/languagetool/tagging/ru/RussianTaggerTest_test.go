package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/tagging/ru/RussianTaggerTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

// Port of RussianTaggerTest.testDictionary — MapWordTagger smoke
func TestRussianTagger_Dictionary(t *testing.T) {
	wt := tagging.MapWordTagger{
		"дом": {tagging.NewTaggedWord("дом", "NN:Inanim:Nom:Masc:Sin")},
	}
	tagger := NewRussianTagger(wt)
	require.Equal(t, RussianDictPath, tagger.GetDictionaryPath())
	got := tagger.TagWord("дом")
	require.Len(t, got, 1)
	require.Equal(t, "NN:Inanim:Nom:Masc:Sin", got[0].PosTag)
}

// Port of RussianTaggerTest.testTagger
func TestRussianTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"это":  {tagging.NewTaggedWord("это", "PNdem")},
		"тест": {tagging.NewTaggedWord("тест", "NN")},
	}
	got := NewRussianTagger(wt).Tag([]string{"Это", "тест", "xyz"})
	require.Len(t, got, 3)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.NotNil(t, got[1].GetReadings()[0].GetPOSTag())
	require.Nil(t, got[2].GetReadings()[0].GetPOSTag())
}
