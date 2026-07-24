package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseAgreementReplaceRule(t *testing.T) {
	rule := NewPortugueseAgreementReplaceRule(nil)
	// Java example: abstracto → abstrato
	matches := rule.Match(languagetool.AnalyzePlain("um conceito abstracto"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "abstrato", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("um conceito abstrato"))))
}

func TestPortugueseReplaceRule_Loads(t *testing.T) {
	// Dictionary may be empty (comments only); load must not panic.
	_ = NewPortugueseReplaceRule(nil)
}
