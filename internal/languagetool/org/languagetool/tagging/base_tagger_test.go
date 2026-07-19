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

// Java getAnalyzedTokens: Title/ALLCAPS also load lowercase tags (not mixed case).
func TestBaseTagger_MergesLowercaseForTitleCase(t *testing.T) {
	m := MapWordTagger{
		"house": {NewTaggedWord("house", "NN")},
		// no "House" entry
	}
	bt := NewBaseTagger(m, "/en/dict", "en", true)
	// exact empty, lower has tags → merge for Title case
	got := bt.TagWord("House")
	require.NotEmpty(t, got)
	require.Equal(t, "NN", got[0].PosTag)
	// exact-only must not invent merge
	require.Empty(t, bt.TagWordExact("House"))
}

// Java: mixed case does not get lowercase tags.
func TestBaseTagger_NoMergeForMixedCase(t *testing.T) {
	m := MapWordTagger{
		"ipod": {NewTaggedWord("ipod", "NN")},
	}
	bt := NewBaseTagger(m, "/en/dict", "en", true)
	require.Empty(t, bt.TagWord("iPod"))
}

// Java tagLowercaseWithUppercase: only UppercaseFirstChar, not full ToUpper.
func TestBaseTagger_LowercaseWithUppercaseFirst(t *testing.T) {
	m := MapWordTagger{
		"Dog": {NewTaggedWord("dog", "NN")},
	}
	bt := NewBaseTagger(m, "/en/dict", "en", true)
	got := bt.TagWord("dog") // empty exact+lower; try "Dog"
	require.NotEmpty(t, got)
	// full TOUPPER invent must not run
	m2 := MapWordTagger{"DOG": {NewTaggedWord("dog", "NN")}}
	bt2 := NewBaseTagger(m2, "/en/dict", "en", true)
	require.Empty(t, bt2.TagWord("dog"), "Java does not try full ToUpper")
}
