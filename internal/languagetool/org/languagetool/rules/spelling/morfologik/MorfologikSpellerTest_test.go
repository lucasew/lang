package morfologik

// Twin of languagetool-core/src/test/java/org/languagetool/rules/spelling/morfologik/MorfologikSpellerTest.java
// Uses official /xx/spelling/test.dict FSA (inspiration languagetool-core test resources) — not map invent.

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const javaTestDictClasspath = "/xx/spelling/test.dict"

// testDictSpeller constructs like Java: new MorfologikSpeller("/xx/spelling/test.dict"[, maxEdit]).
// Requires official binary on disk (DiscoverLanguageDict → languagetool-core test resources).
func testDictSpeller(t *testing.T, maxEdit int) *MorfologikSpeller {
	t.Helper()
	path := DiscoverLanguageDict(javaTestDictClasspath)
	require.NotEmpty(t, path, "official %s not found under inspiration/languagetool", javaTestDictClasspath)
	sp := NewMorfologikSpeller(javaTestDictClasspath, maxEdit)
	// Java MorfologikSpeller(path) loads Dictionary in the constructor; Go attaches the
	// same discovered FSA explicitly (global New auto-attach would break map-inject tests).
	require.True(t, sp.AttachBinaryDictionary(path), "AttachBinaryDictionary(%s)", path)
	require.True(t, sp.HasDictionary(), "binary FSA required for twin")
	require.Equal(t, path, sp.BinaryDictPath)
	require.Empty(t, sp.Words, "twin must not use map invent for membership")
	return sp
}

// weightedListString ports Java List<WeightedSuggestion>.toString() form: [wordone/51].
func weightedListString(ws []WeightedSuggestion) string {
	if len(ws) == 0 {
		return "[]"
	}
	parts := make([]string, len(ws))
	for i, w := range ws {
		parts[i] = w.String()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// Port of MorfologikSpellerTest.testIsMisspelled
func TestMorfologikSpeller_IsMisspelled(t *testing.T) {
	sp := testDictSpeller(t, 1)
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
// Java getSuggestions returns List<WeightedSuggestion>; list.toString is [wordone/51].
func TestMorfologikSpeller_GetSuggestions(t *testing.T) {
	sp1 := testDictSpeller(t, 1)
	sp2 := testDictSpeller(t, 2)

	require.Equal(t, "[]", weightedListString(sp1.GetWeightedSuggestions("wordone")))
	require.Equal(t, "[wordone/51]", weightedListString(sp1.GetWeightedSuggestions("wordonex")))
	require.Equal(t, "[]", weightedListString(sp2.GetWeightedSuggestions("wordone")))
	require.Equal(t, "[wordone/51]", weightedListString(sp2.GetWeightedSuggestions("wordonex")))

	require.Equal(t, "[]", weightedListString(sp1.GetWeightedSuggestions("wordonix")))
	require.Equal(t, "[wordone/77]", weightedListString(sp2.GetWeightedSuggestions("wordonix")))

	require.Equal(t, "[]", weightedListString(sp2.GetWeightedSuggestions("wordoxix")))
}

// Explicit weight fields (same outcomes as Java WeightedSuggestion toString on test.dict).
func TestMorfologikSpeller_GetWeightedSuggestions_JavaWeights(t *testing.T) {
	sp1 := testDictSpeller(t, 1)
	sp2 := testDictSpeller(t, 2)

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
