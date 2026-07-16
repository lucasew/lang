package km

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestKhmerSimpleReplaceRule(t *testing.T) {
	rule := NewKhmerSimpleReplaceRule(nil)
	// From coherency.txt: សំដី=សម្ដី
	matches := rule.Match(languagetool.AnalyzePlain("សំដី"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "សម្ដី", matches[0].GetSuggestedReplacements()[0])
}
