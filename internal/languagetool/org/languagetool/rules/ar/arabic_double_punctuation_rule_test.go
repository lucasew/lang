package ar

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicDoublePunctuationRule(t *testing.T) {
	rule := NewArabicDoublePunctuationRule(nil)
	require.Equal(t, "ARABIC_DOUBLE_PUNCTUATION", rule.GetID())
	// AnalyzePlain glues "،،" into one token; space-separated commas match Java token path.
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("نعم ، ،"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("نعم ،"))))
}
