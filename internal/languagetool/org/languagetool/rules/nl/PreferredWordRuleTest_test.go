package nl

// Twin of languagetool-language-modules/nl/src/test/java/org/languagetool/rules/nl/PreferredWordRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPreferredWordRule_Rule(t *testing.T) {
	rule := NewPreferredWordRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("rijwiel"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "fiets", matches[0].GetSuggestedReplacements()[0])
}
