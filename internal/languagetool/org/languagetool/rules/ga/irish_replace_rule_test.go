package ga

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestIrishReplaceRule(t *testing.T) {
	rule := NewIrishReplaceRule(nil)
	// Example from Java: bhúr → bhur
	matches := rule.Match(languagetool.AnalyzePlain("ar bhúr gcuid cainte."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "bhur", matches[0].GetSuggestedReplacements()[0])
}
