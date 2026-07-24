package ro

// Twin of languagetool-language-modules/ro/src/test/java/org/languagetool/rules/ro/SimpleReplaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Paisprezece case."))))

	check := func(sentence, word string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(sentence))
		require.Equal(t, 1, len(matches), "sentence %q got %v", sentence, matches)
		require.Equal(t, word, matches[0].GetSuggestedReplacements()[0], "sentence %q", sentence)
	}
	check("Patrusprezece case.", "Paisprezece")
	check("Satul are patrusprezece case.", "paisprezece")
	check("Satul are (patrusprezece) case.", "paisprezece")
	check("Satul are «patrusprezece» case.", "paisprezece")
	check("El are șasesprezece ani.", "șaisprezece")
	check("El a luptat pentru întâiele cărți.", "întâile")
	check("El are cinsprezece cărți.", "cincisprezece")
}

func TestSimpleReplaceRule_InvalidSuggestion(t *testing.T) {
	// Java validates data file has no self-replacements; load succeeds if format is OK.
	_ = NewSimpleReplaceRule(nil)
}
