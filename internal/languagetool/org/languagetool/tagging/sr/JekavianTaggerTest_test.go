package sr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestJekavianTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{"svijet": {tagging.NewTaggedWord("svijet", "N")}}
	tagger := NewJekavianTagger(wt)
	got := tagger.TagWord("svijet")
	require.Len(t, got, 1)
}

func TestJekavianTagger_Dictionary(t *testing.T) {
	require.Equal(t, JekavianDictionaryPath, NewJekavianTagger(nil).GetDictionaryPath())
}
