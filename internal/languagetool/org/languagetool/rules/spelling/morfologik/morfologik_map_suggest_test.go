package morfologik

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Map-path plain/user dict: AddWord → ensureWordsAsDictionary (FSABuilder+CFSA2+Speller),
// not invent map Levenshtein scan (Java MultiSpeller.getDictionary runtime FSA).
func TestMapWordsSuggest_Transposition(t *testing.T) {
	sp := NewMorfologikSpeller("/xx/map.dict", 1)
	sp.AddWord("receive")
	sp.AddWord("recipe")
	// recieve → receive is adjacent transposition of e/i (Damerau distance 1)
	sugs := sp.FindReplacements("recieve")
	require.Contains(t, sugs, "receive", "sugs=%v", sugs)
	// weight uses dist 1
	ws := sp.GetWeightedSuggestions("recieve")
	require.NotEmpty(t, ws)
	for _, w := range ws {
		if w.Word == "receive" {
			// dist 1, freq 0 → 1*26+26-0-1 = 51
			require.Equal(t, 51, w.Weight)
		}
	}
	// After suggest, Words map became a CFSA2 Dictionary (Java runtime dict path).
	require.NotNil(t, sp.binaryDict)
}

func TestMapWordsSuggest_Edit2NeedsMaxEdit2(t *testing.T) {
	sp1 := NewMorfologikSpeller("/xx/map.dict", 1)
	sp1.AddWord("guarantee")
	// garentee is classic edit-2 from guarantee
	require.NotContains(t, sp1.FindReplacements("garentee"), "guarantee")

	sp2 := NewMorfologikSpeller("/xx/map.dict", 2)
	sp2.AddWord("guarantee")
	require.Contains(t, sp2.FindReplacements("garentee"), "guarantee")
}

// Java MorfologikSpeller.getSuggestions builds new Speller each call (HMatrix comment);
// findRepl + run-on share that Speller instance.
func TestGetWeightedSuggestions_RunOnWithMapWords(t *testing.T) {
	sp := NewMorfologikSpeller("/xx/map.dict", 1)
	sp.AddWord("the")
	sp.AddWord("cat")
	ws := sp.GetWeightedSuggestions("thecat")
	found := false
	for _, w := range ws {
		if w.Word == "the cat" {
			found = true
			// run-on CandidateData dist 1 → weight 1*26+26-0-1 = 51
			require.Equal(t, 51, w.Weight)
		}
	}
	require.True(t, found, "expected run-on the cat, got %+v", ws)
}
