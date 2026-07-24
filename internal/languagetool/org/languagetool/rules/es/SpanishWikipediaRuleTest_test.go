package es

// Twin of languagetool-language-modules/es/src/test/java/org/languagetool/rules/es/SpanishWikipediaRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSpanishWikipediaRule_Rule(t *testing.T) {
	rule := NewSpanishWikipediaRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Estas frases no tienen errores frecuentes en la Wikipedia."))))

	check := func(sentence, word string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(sentence))
		require.Equal(t, 1, len(matches), "sentence %q", sentence)
		require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
		require.Equal(t, word, matches[0].GetSuggestedReplacements()[0])
	}
	// beginning of a sentence
	check("Sucedió ayer. Murio sin que nadie lo esperase.", "Murió")
	check("Murio sin que nadie lo esperase.", "Murió")
	// inside sentence
	check("Ayer murio sin que nadie lo esperase.", "murió")
}
