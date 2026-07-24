package morfologik

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of morfologik.speller.Speller.replaceRunOnWordCandidates /
// MorfologikSpeller.getSuggestions run-on arm.

func TestReplaceRunOnWordCandidates_MapInject(t *testing.T) {
	sp := NewMorfologikSpeller("/xx/test.dict", 1)
	for _, w := range []string{"the", "cat", "great", "elephant"} {
		sp.AddWord(w)
	}
	// misspelled run-on
	cands := sp.ReplaceRunOnWordCandidates("thecat")
	require.NotEmpty(t, cands)
	found := false
	for _, c := range cands {
		if c.Word == "the cat" {
			found = true
			// distance 1 → weight 26+26-0-1 = 51 with freq 0
			require.Equal(t, 51, c.Weight)
		}
	}
	require.True(t, found, "cands=%v", cands)

	// known word → no run-on
	require.Empty(t, sp.ReplaceRunOnWordCandidates("the"))

	// SupportRunOnWords false
	sp.SupportRunOnWords = false
	require.Empty(t, sp.ReplaceRunOnWordCandidates("thecat"))
}

func TestReplaceRunOnWordCandidates_SentenceStartPrefix(t *testing.T) {
	sp := NewMorfologikSpeller("/xx/test.dict", 1)
	sp.AddWord("the")
	sp.AddWord("cat")
	// "Thecat" — prefix "The" accepted via lower "the"
	cands := sp.ReplaceRunOnWordCandidates("Thecat")
	require.Contains(t, sp.ReplaceRunOnWords("Thecat"), "The cat")
	require.NotEmpty(t, cands)
}

func TestGetWeightedSuggestions_IncludesRunOn(t *testing.T) {
	sp := NewMorfologikSpeller("/xx/test.dict", 1)
	for _, w := range []string{"the", "cat", "that"} {
		sp.AddWord(w)
	}
	// no edit-1 peers for "thecat" among small map? "that" is dist far; run-on still present
	sugs := sp.GetWeightedSuggestions("thecat")
	found := false
	for _, s := range sugs {
		if s.Word == "the cat" {
			found = true
		}
	}
	require.True(t, found, "sugs=%v", sugs)
}

func TestBinaryRunOn_EN(t *testing.T) {
	path := DiscoverLanguageDict("/en/hunspell/en_US.dict")
	if path == "" {
		t.Skip("en_US.dict not in tree")
	}
	sp := NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	require.True(t, sp.AttachBinaryDictionary(path))
	require.True(t, sp.SupportRunOnWords)
	// classic run-on if both sides in dict
	words := sp.ReplaceRunOnWords("thecat")
	require.Contains(t, words, "the cat", "words=%v", words)
}
