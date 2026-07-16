package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceRule(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Please disencourage that behavior."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "discourage", matches[0].GetSuggestedReplacements()[0])

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Please discourage that behavior."))))
}
