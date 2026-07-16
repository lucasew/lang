package es

// Twin of SimpleReplaceVerbsRuleTest — surface keys only.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceVerbsRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceVerbsRule(nil)
	// eruptar is a dictionary key
	matches := rule.Match(languagetool.AnalyzePlain("Puede eruptar el volcán."))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements()[0], "eructar")
}
