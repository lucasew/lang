package ca

// Twin of SimpleReplaceVerbsRuleTest — surface keys only.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceVerbsRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceVerbsRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Va retumbar fort."))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements()[0], "retrunyir")
}
