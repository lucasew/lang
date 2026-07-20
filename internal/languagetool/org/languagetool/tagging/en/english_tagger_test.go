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

// Java BaseTagger: pos += word.length() (UTF-16). Non-BMP emoji advances by 2.
func TestEnglishTagger_StartPosUTF16(t *testing.T) {
	wt := tagging.MapWordTagger{
		"ok": {tagging.NewTaggedWord("ok", "JJ")},
	}
	tagger := NewEnglishTagger(wt)
	// 😀 = 1 code point, 2 UTF-16 units; then "ok"
	got := tagger.Tag([]string{"😀", "ok"})
	require.Len(t, got, 2)
	require.Equal(t, 0, got[0].GetStartPos())
	require.Equal(t, 2, got[1].GetStartPos(), "Java word.length() for emoji is 2 UTF-16 units")
	require.Equal(t, 2, tagging.UTF16Len("😀"))
}
