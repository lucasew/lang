package tagging

// Twin of MorfologikTaggerTest — injected Lookup stand-in for binary dict.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikTagger_Tag(t *testing.T) {
	dict := map[string][]TaggedWord{
		"cars": {NewTaggedWord("car", "NNS")},
		"car":  {NewTaggedWord("car", "NN")},
	}
	tagger := NewMorfologikTaggerWithLookup(func(word string) []TaggedWord {
		return dict[word]
	})
	require.Empty(t, tagger.Tag("nosuchword"))
	got := tagger.Tag("cars")
	require.Len(t, got, 1)
	require.Equal(t, "car", got[0].GetLemma())
	require.Equal(t, "NNS", got[0].GetPosTag())

	tagger.SetInternTags(true)
	got2 := tagger.Tag("car")
	require.Equal(t, "NN", got2[0].GetPosTag())
}

func TestMorfologikTagger_PositionWithIgnoredChars(t *testing.T) {
	// Soft surface: Tag returns readings independent of positions (positions are sentence-level).
	tagger := NewMorfologikTaggerWithLookup(func(word string) []TaggedWord {
		if word == "foo" {
			return []TaggedWord{NewTaggedWord("foo", "X")}
		}
		return nil
	})
	require.NotNil(t, tagger.Tag("foo"))
	require.Empty(t, tagger.Tag(""))
}
