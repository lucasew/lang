package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUkrainianCommaWhitespaceRule(t *testing.T) {
	rule := NewUkrainianCommaWhitespaceRule(nil)
	// em dash with surrounding spaces should not use right-bracket path; just smoke
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Слово — слово."))))
	// space before comma still flagged
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Слово , слово."))))
}
