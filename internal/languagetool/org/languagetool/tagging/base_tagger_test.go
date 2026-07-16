package tagging

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBaseTagger(t *testing.T) {
	m := MapWordTagger{
		"dog": {NewTaggedWord("dog", "NN")},
		"Dog": {NewTaggedWord("dog", "NN")},
	}
	bt := NewBaseTagger(m, "/en/dict", "en", true)
	require.Equal(t, "/en/dict", bt.GetDictionaryPath())
	require.Contains(t, bt.GetManualAdditionsFileNames()[0], "added.txt")
	tags := bt.TagWords([]string{"dog", "unknown"})
	require.Len(t, tags, 2)
	require.Equal(t, "NN", tags[0][0].PosTag)
	require.Empty(t, tags[1])
	// lowercase form is present in map
	require.NotEmpty(t, bt.TagWord("dog"))
}
