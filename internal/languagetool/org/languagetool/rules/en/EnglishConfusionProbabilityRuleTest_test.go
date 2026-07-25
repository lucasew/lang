package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/stretchr/testify/require"
)

// TestEnglishConfusionProbabilityRule_Constructor twins
// EnglishConfusionProbabilityRuleTest.testConstructor:
// constructing with FakeLanguageModel + English does not throw; id CONFUSION_RULE
// (Java EnglishConfusionProbabilityRule does not override getId).
// Java @Ignore testRule (Lucene ngram index) is intentionally not ported.
func TestEnglishConfusionProbabilityRule_Constructor(t *testing.T) {
	// Java: new EnglishConfusionProbabilityRule(TestTools.getEnglishMessages(),
	//   new FakeLanguageModel(), Languages.getLanguageForShortCode("en"));
	lm := ngrams.UniformLanguageModel(0.5, 1) // FakeLanguageModel stand-in
	r := NewEnglishConfusionProbabilityRule(lm)
	require.NotNil(t, r)
	require.Equal(t, ngrams.ConfusionRuleID, r.GetID())
	require.Equal(t, "CONFUSION_RULE", r.GetID())

	// Java-visible constructor meta (not invent)
	require.Equal(t, []string{"confusion_sets.txt"}, r.GetFilenames())
	require.Len(t, r.Exceptions, 483)
	require.Len(t, EnglishConfusionAntiPatterns, 32)
	require.NotNil(t, r.IsException)
	require.NotNil(t, r.IsCoveredByAntiPattern)

	// Example pair breaks/brakes
	inex := r.GetIncorrectExamples()
	require.NotEmpty(t, inex)
	require.Contains(t, inex[0].GetExample(), "breaks")
	cex := r.GetCorrectExamples()
	require.NotEmpty(t, cex)
	require.Contains(t, cex[0].GetExample(), "brakes")

	// Official confusion sets loaded (getFilenames path); empty only if resource absent
	if discoverENL2ConfusionSet("confusion_sets.txt") != "" {
		require.NotEmpty(t, r.WordToPairs, "constructor must load confusion_sets.txt like Java CPR")
	}
}
