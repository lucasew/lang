package ar

// Twin of ArabicConfusionProbabilityRuleTest.
// Java only has testConstructor (FakeLanguageModel + ar); extra cases assert
// Java-visible constructor meta (id, filenames, example pair, official confusion_sets).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/stretchr/testify/require"
)

// TestArabicConfusionProbabilityRule_Constructor twins
// ArabicConfusionProbabilityRuleTest.testConstructor:
// constructing with FakeLanguageModel + ar does not throw; id CONFUSION_RULE
// (Java ArabicConfusionProbabilityRule does not override getId).
func TestArabicConfusionProbabilityRule_Constructor(t *testing.T) {
	// Java: new ArabicConfusionProbabilityRule(TestTools.getEnglishMessages(),
	//   new FakeLanguageModel(), Languages.getLanguageForShortCode("ar"));
	lm := ngrams.UniformLanguageModel(0.5, 1) // FakeLanguageModel stand-in
	r := NewArabicConfusionProbabilityRule(lm)
	require.NotNil(t, r)
	require.NotNil(t, r.ConfusionProbabilityRule)
	require.Equal(t, ngrams.ConfusionRuleID, r.GetID())
	require.Equal(t, "CONFUSION_RULE", r.GetID())

	// Java-visible constructor meta (not invent): no override of getFilenames
	require.Equal(t, []string{"confusion_sets.txt"}, r.GetFilenames())
	require.Equal(t, ArabicConfusionFilenames(), r.GetFilenames())

	// Example pair الضن → الظن (trailing unclosed <marker> kept as upstream Java strings)
	inex := r.GetIncorrectExamples()
	require.Len(t, inex, 1)
	require.Equal(t, "إن بعض <marker>الضن</marker> إثم.<marker>", inex[0].GetExample())
	require.Equal(t, []string{"الظن"}, inex[0].GetCorrections())
	cex := r.GetCorrectExamples()
	require.Len(t, cex, 1)
	require.Equal(t, "إن بعض <marker>الظن</marker> إثم.<marker>", cex[0].GetExample())

	// Official /ar/confusion_sets.txt is header comments only → empty WordToPairs after parse.
	// Do not invent الضن/الظن data rows even though the example pair mentions them.
	require.Empty(t, r.WordToPairs, "header-only confusion_sets.txt must yield empty WordToPairs")
	require.Len(t, r.WordToPairs, 0)
}
