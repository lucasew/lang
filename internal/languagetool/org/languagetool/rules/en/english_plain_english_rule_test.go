package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishPlainEnglishRule(t *testing.T) {
	rule := NewEnglishPlainEnglishRule(nil)
	// Java example: fatal outcome → death
	matches := rule.Match(languagetool.AnalyzePlain("a fatal outcome occurred"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "death", matches[0].GetSuggestedReplacements()[0])
}
