package morfologik

// Twin of languagetool-core/src/test/java/org/languagetool/rules/spelling/morfologik/MorfologikSpellerTest.java
// Binary test.dict deferred — map inject mirrors Java /xx/spelling/test.dict cases.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testDictSpeller(maxEdit int) *MorfologikSpeller {
	sp := NewMorfologikSpeller("/xx/spelling/test.dict", maxEdit)
	for _, w := range []string{"wordone", "wordtwo", "Uppercase", "Häuser"} {
		sp.AddWord(w)
	}
	return sp
}

// Port of MorfologikSpellerTest.testIsMisspelled
func TestMorfologikSpeller_IsMisspelled(t *testing.T) {
	sp := testDictSpeller(1)
	require.True(t, sp.ConvertsCase())

	require.False(t, sp.IsMisspelled("wordone"))
	require.False(t, sp.IsMisspelled("Wordone"))
	require.False(t, sp.IsMisspelled("wordtwo"))
	require.False(t, sp.IsMisspelled("Wordtwo"))
	require.False(t, sp.IsMisspelled("Uppercase"))
	require.False(t, sp.IsMisspelled("Häuser"))

	require.True(t, sp.IsMisspelled("Hauser"))
	require.True(t, sp.IsMisspelled("wordones"))
	require.True(t, sp.IsMisspelled("nosuchword"))
}

// Port of MorfologikSpellerTest.testGetSuggestions
func TestMorfologikSpeller_GetSuggestions(t *testing.T) {
	sp1 := testDictSpeller(1)
	sp2 := testDictSpeller(2)

	// exact dictionary form → no suggestion needed
	require.Empty(t, sp1.GetSuggestions("wordone"))
	require.Empty(t, sp2.GetSuggestions("wordone"))

	// single edit: wordonex → wordone at distance 1
	require.Contains(t, sp1.GetSuggestions("wordonex"), "wordone")
	require.Contains(t, sp2.GetSuggestions("wordonex"), "wordone")

	// two edits: wordonix needs distance 2
	require.Empty(t, sp1.GetSuggestions("wordonix"))
	require.Contains(t, sp2.GetSuggestions("wordonix"), "wordone")

	// too far
	require.Empty(t, sp2.GetSuggestions("wordoxix"))
}

// Twin of Java WeightedSuggestion toString weights on test.dict (freq 0):
// wordonex → wordone/51 (dist1*26+26-0-1); wordonix → wordone/77 (dist2).
func TestMorfologikSpeller_GetWeightedSuggestions_JavaWeights(t *testing.T) {
	sp1 := testDictSpeller(1)
	sp2 := testDictSpeller(2)

	require.Empty(t, sp1.GetWeightedSuggestions("wordone"))

	w1 := sp1.GetWeightedSuggestions("wordonex")
	require.NotEmpty(t, w1)
	require.Equal(t, "wordone", w1[0].Word)
	require.Equal(t, 51, w1[0].Weight, "Java: wordone/51")
	require.Equal(t, "wordone/51", w1[0].String())

	require.Empty(t, sp1.GetWeightedSuggestions("wordonix"))
	w2 := sp2.GetWeightedSuggestions("wordonix")
	require.NotEmpty(t, w2)
	require.Equal(t, "wordone", w2[0].Word)
	require.Equal(t, 77, w2[0].Weight, "Java: wordone/77")
	require.Equal(t, "wordone/77", w2[0].String())

	require.Empty(t, sp2.GetWeightedSuggestions("wordoxix"))
}
