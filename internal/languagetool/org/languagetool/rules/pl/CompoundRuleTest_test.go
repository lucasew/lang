package pl

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/rules/pl/CompoundRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCompoundRule_Rule(t *testing.T) {
	rule := NewCompoundRule(nil)
	check := func(expectedErrors int, text string, expSuggestions ...string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(text))
		require.Equal(t, expectedErrors, len(matches), "text %q got %v", text, matches)
		if len(expSuggestions) > 0 {
			require.Equal(t, 1, expectedErrors)
			require.Equal(t, expSuggestions, matches[0].GetSuggestedReplacements())
		}
	}
	check(0, "Nie róbmy nic na łapu-capu.")
	check(0, "Jedzmy kogel-mogel.")
	check(1, "bim bom", "bim-bom")
}

// Port of CompoundRuleTest.testCompoundFile — data load + rule surface (full file matrix deferred).
func TestCompoundRule_CompoundFile(t *testing.T) {
	data := loadCompoundData()
	require.NotNil(t, data)
	require.NotEmpty(t, data.IncorrectCompounds)
	// known entry from compounds.txt used by testRule
	rule := NewCompoundRule(nil)
	require.Equal(t, "PL_COMPOUNDS", rule.GetID())
	// file-backed pair still matches
	matches := rule.Match(languagetool.AnalyzePlain("bim bom"))
	require.Len(t, matches, 1)
	require.Contains(t, matches[0].GetSuggestedReplacements(), "bim-bom")
}
