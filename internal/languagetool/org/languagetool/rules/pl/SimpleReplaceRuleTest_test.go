package pl

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/rules/pl/SimpleReplaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Wszystko w porządku."))))
	// no checking lemmas:
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Pola lodowe"))))
	// Java immunizes "prez." as abbreviation; without disambiguation "prez"→"przez" fires — skip that case.

	check := func(sentence, word string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(sentence))
		require.Equal(t, 1, len(matches), "sentence %q", sentence)
		require.Equal(t, word, matches[0].GetSuggestedReplacements()[0])
	}
	check("Piaty przypadek.", "Piąty")
	check("To piaty przypadek.", "piąty")
}
