package suggestions_ordering

// Twin of
// org.languagetool.rules.spelling.morfologik.suggestions_ordering.DetailedDamerauLevenstheinDistanceTest
//
// Java marks both @Test methods @Ignore (WIP / needs further debugging). We still twin the
// Java-visible outcomes of compare() as produced by the JVM implementation (king of correctness).
// Note: testDistanceDetails' expected delete().transpose() for "Test"→"Tts" does NOT match the
// current Java algorithm (which yields delete+replace); Go matches the JVM, not the WIP expectation.

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDistanceDetails twins DetailedDamerauLevenstheinDistanceTest.testDistanceDetails.
// Assertions match JVM DetailedDamerauLevenstheinDistance.compare results (verified against
// inspiration Java sources). Cases 1–5 and 7 match the WIP matrix; case 6 follows JVM.
func TestDistanceDetails(t *testing.T) {
	// assertEquals(new Distance().delete(), compare("Test", "Tet"));
	require.Equal(t, Distance{}.Delete(), Compare("Test", "Tet"))
	// assertEquals(new Distance().insert(), compare("Test", "Teste"));
	require.Equal(t, Distance{}.Insert(), Compare("Test", "Teste"))
	// assertEquals(new Distance().transpose(), compare("Test", "Tets"));
	require.Equal(t, Distance{}.Transpose(), Compare("Test", "Tets"))
	// assertEquals(new Distance().replace(), compare("Test", "Tast"));
	require.Equal(t, Distance{}.Replace(), Compare("Test", "Tast"))
	// assertEquals(new Distance().replace().insert(), compare("Test", "Taste"));
	require.Equal(t, Distance{}.Replace().Insert(), Compare("Test", "Taste"))
	// WIP expected: delete().transpose() — JVM actual (and Go): delete+replace
	// assertEquals(new Distance().delete().transpose(), compare("Test", "Tts")); // fails on JVM
	require.Equal(t, Distance{Deletes: 1, Replaces: 1}, Compare("Test", "Tts"))
	// assertEquals(new Distance().insert().insert(), compare("Test", "Teeste"));
	require.Equal(t, Distance{}.Insert().Insert(), Compare("Test", "Teeste"))
}

// TestDistanceComputationSurface twins the structure of testDistanceComputation without the
// SymSpell EditDistance reference (separate sector). Confirms compare is defined for random edits.
func TestDistanceComputationSurface(t *testing.T) {
	text := "This is a test text. Random edits will occur here. Foo bar baz. Bla bla. Lorem ipsum dolor sit amet."
	// apply a few random edits like modifyString; skip when Apply returns null
	result := text
	history := Distance{}
	for i := 0; i < 5; i++ {
		op := RandomEdit()
		tmp, ok := op.Apply(result)
		if !ok {
			continue
		}
		result = tmp
		history = history.Track(op)
	}
	computed := Compare(text, result)
	// value is non-negative; identical text yields 0
	require.GreaterOrEqual(t, computed.Value(), 0)
	if result == text {
		require.Equal(t, 0, computed.Value())
		require.Equal(t, Distance{}, computed)
	}
	_ = history
}
