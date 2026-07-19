package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestEnglishTagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"dogs": {tagging.NewTaggedWord("dog", "NNS")},
		"dog":  {tagging.NewTaggedWord("dog", "NN")},
	}
	tagger := NewEnglishTagger(wt)
	got := tagger.Tag([]string{"dogs", "DOGS", "xyz"})
	require.Len(t, got, 3)
	require.Equal(t, "dogs", got[0].GetToken())
	// all upper falls back to lower/first upper
	require.NotEmpty(t, got[1].GetReadings())
	// unknown
	require.Equal(t, "xyz", got[2].GetToken())
}

func TestEnglishTagger_TypographicApostrophe(t *testing.T) {
	wt := tagging.MapWordTagger{"don't": {tagging.NewTaggedWord("do", "VBP")}}
	tagger := NewEnglishTagger(wt)
	got := tagger.Tag([]string{"don’t"})
	require.Len(t, got, 1)
	require.True(t, got[0].HasTypographicApostrophe())
}
