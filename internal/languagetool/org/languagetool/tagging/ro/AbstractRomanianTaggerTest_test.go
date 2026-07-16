package ro

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestAbstractRomanianTagger_Dictionary(t *testing.T) {
	wt := tagging.MapWordTagger{"carte": {tagging.NewTaggedWord("carte", "S")}}
	tagger := NewRomanianTagger(wt)
	require.Equal(t, RomanianDictPath, tagger.GetDictionaryPath())
	require.Len(t, tagger.TagWord("carte"), 1)
}
