package ar

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicDiacriticsRule(t *testing.T) {
	rule := NewArabicDiacriticsRule(nil)
	// Example from Java: تجربة → تجرِبة
	matches := rule.Match(languagetool.AnalyzePlain("تجربة"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "تجرِبة", matches[0].GetSuggestedReplacements()[0])
}
