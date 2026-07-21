package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestEnglishTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"house": {tagging.NewTaggedWord("house", "N")},
	}
	tagger := NewEnglishTagger(wt)
	got := tagger.Tag([]string{"house", "xyz"})
	require.Len(t, got, 2)
	require.NotEmpty(t, got[0].GetReadings())
	// unknown word still yields a reading
	require.NotEmpty(t, got[1].GetReadings())
}

func TestEnglishTagger_DictionaryPath(t *testing.T) {
	tagger := NewEnglishTagger(nil)
	require.NotEmpty(t, tagger.GetDictionaryPath())
}

// Twin of EnglishTaggerTest.testDictionary — path/dict availability (full morfologik scan is N/A without dict bytes).
func TestEnglishTagger_Dictionary(t *testing.T) {
	tagger := NewEnglishTagger(nil)
	// Java TestTools.testDictionary walks the dict; we assert resource path wiring exists.
	require.NotEmpty(t, tagger.GetDictionaryPath())
}

// Twin of EnglishTaggerTest.testLemma — morph with map tagger (no invent readings).
func TestEnglishTagger_Lemma(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Trump": {
			tagging.NewTaggedWord("Trump", "NNP"),
			tagging.NewTaggedWord("trump", "NN"),
			tagging.NewTaggedWord("trump", "VB"),
			tagging.NewTaggedWord("trump", "VBP"),
		},
		"works": {
			tagging.NewTaggedWord("work", "VBZ"),
			tagging.NewTaggedWord("works", "NNS"),
		},
	}
	got := NewEnglishTagger(wt).Tag([]string{"Trump", "works"})
	require.Len(t, got, 2)
	require.Len(t, got[0].GetReadings(), 4)
	require.Len(t, got[1].GetReadings(), 2)
}
