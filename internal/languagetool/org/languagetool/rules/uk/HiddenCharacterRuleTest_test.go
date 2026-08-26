package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/HiddenCharacterRuleTest.java
// Analysis: AnalyzePlain — rule only inspects token surfaces for U+00AD; WordTokenizer
// keeps soft hyphen in the token (Java end-state getToken() also has orig soft-hyphen
// after addReading; for this surface-only rule plain analyze is sufficient).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestHiddenCharacterRule_Rule(t *testing.T) {
	rule := NewHiddenCharacterRule(nil)
	require.Equal(t, "UK_HIDDEN_CHARS", rule.GetID())
	require.Equal(t, "Приховані символи: знак м’якого перенесення", rule.GetDescription())
	require.Equal(t, "Приховані символи", rule.GetShort())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.CatMisc.GetCategory(nil).GetID(), rule.GetCategory().GetID())

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("сміття"))))

	// incorrect sentences:
	matches := rule.Match(languagetool.AnalyzePlain("смі\u00ADття"))
	// check match positions:
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"сміття"}, matches[0].GetSuggestedReplacements())

	// Java-visible message / short / span (extra fidelity; not invent)
	tok := "смі\u00ADття"
	wantMsg := tok + rule.GetSuggestion(tok) + "сміття"
	require.Equal(t, wantMsg, matches[0].GetMessage())
	// getSuggestion highlights soft hyphen as '-'
	require.Contains(t, rule.GetSuggestion(tok), "«смі-ття»")
	require.Equal(t, "Приховані символи", matches[0].GetShortMessage())
	// whole token span under AnalyzePlain (UTF-16: 7 units)
	require.Equal(t, 0, matches[0].FromPos)
	require.Equal(t, 7, matches[0].ToPos)
	// replacement has no soft hyphen
	require.False(t, strings.ContainsRune(matches[0].GetSuggestedReplacements()[0], HIDDEN_CHAR))
}
