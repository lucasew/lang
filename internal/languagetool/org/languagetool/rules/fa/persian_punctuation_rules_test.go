package fa

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPersianDoublePunctuationRule(t *testing.T) {
	rule := NewPersianDoublePunctuationRule(nil)
	require.Equal(t, "PERSIAN_DOUBLE_PUNCTUATION", rule.GetID())
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("بله ، ،"))))
}

func TestPersianCommaWhitespaceRule(t *testing.T) {
	rule := NewPersianCommaWhitespaceRule(nil)
	require.Equal(t, "PERSIAN_COMMA_PARENTHESIS_WHITESPACE", rule.GetID())
	// space before Arabic comma
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("بله ،"))))
}
