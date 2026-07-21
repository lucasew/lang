package morfologik

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Map-path plain/user dict: SpellerED (Damerau) not invent plain Levenshtein.
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
