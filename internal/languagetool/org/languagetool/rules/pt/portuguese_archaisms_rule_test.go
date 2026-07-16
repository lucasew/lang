package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseArchaismsRule(t *testing.T) {
	rule := NewPortugueseArchaismsRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("uma câmera digital"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "câmara", matches[0].GetSuggestedReplacements()[0])
}
