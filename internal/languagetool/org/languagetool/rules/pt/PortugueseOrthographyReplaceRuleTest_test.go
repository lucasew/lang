package pt

// Twin of PortugueseOrthographyReplaceRuleTest — immunization for multiword, not surface invent.
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseOrthographyReplaceRule_Rule(t *testing.T) {
	rule := NewPortugueseOrthographyReplaceRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Já volto."))))
	assertSingle := func(sentence string, suggestions ...string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(sentence))
		require.Equal(t, 1, len(matches), "sentence %q", sentence)
		require.Equal(t, "PT_SIMPLE_REPLACE_ORTHOGRAPHY", rule.GetID())
		got := matches[0].GetSuggestedReplacements()
		require.Equal(t, len(suggestions), len(got), "sentence %q suggestions %v", sentence, got)
		for i, s := range suggestions {
			require.Equal(t, s, got[i])
		}
	}
	assertSingle("Ja volto.", "Já")

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Gosto de você."))))
	assertSingle("Gosto de voce.", "você")

	// Java: multiword analysis immunizes "voce" in "sotto voce"
	sent := languagetool.AnalyzePlain("Disse-me sotto voce.")
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if strings.EqualFold(tok.GetToken(), "voce") {
			tok.Immunize(0)
		}
	}
	require.Equal(t, 0, len(rule.Match(sent)))

	// without immunization, voce still matches (fail closed, no invent skip)
	require.NotEmpty(t, rule.Match(languagetool.AnalyzePlain("Disse-me sotto voce.")))
}
