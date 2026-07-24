package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/HiddenCharacterRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestHiddenCharacterRule_Rule(t *testing.T) {
	rule := NewHiddenCharacterRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("сміття"))))

	matches := rule.Match(languagetool.AnalyzePlain("смі\u00ADття"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"сміття"}, matches[0].GetSuggestedReplacements())
}
