package tagging

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikTagger(t *testing.T) {
	mt := NewMorfologikTaggerWithLookup(func(word string) []TaggedWord {
		if word == "dogs" {
			return []TaggedWord{NewTaggedWord("dog", "NNS")}
		}
		return nil
	})
	require.Equal(t, "NNS", mt.Tag("dogs")[0].PosTag)
	require.Empty(t, mt.Tag("unknown"))
	mt.SetInternTags(true)
	require.True(t, mt.GetInternTags())
}
