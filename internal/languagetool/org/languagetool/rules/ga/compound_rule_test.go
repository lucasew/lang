package ga

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCompoundRule(t *testing.T) {
	rule := NewCompoundRule(nil)
	require.Equal(t, "GA_COMPOUNDS", rule.GetID())
	// Java example: mí úsáid → mí-úsáid
	matches := rule.Match(languagetool.AnalyzePlain("Tá mí úsáid fhisiciúil i gceist."))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements()[0], "mí-úsáid")
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Tá mí-úsáid fhisiciúil i gceist."))))
}
