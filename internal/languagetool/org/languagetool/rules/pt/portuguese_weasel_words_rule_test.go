package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseWeaselWordsRule(t *testing.T) {
	rule := NewPortugueseWeaselWordsRule(nil)

	// Example marker from Java: Diz-se
	matches := rule.Match(languagetool.AnalyzePlain("Diz-se que programas gratuitos não têm qualidade."))
	require.Equal(t, 1, len(matches))
	require.NotEmpty(t, matches[0].GetSuggestedReplacements())

	matches = rule.Match(languagetool.AnalyzePlain("Acredita-se que vai chover."))
	require.Equal(t, 1, len(matches))
}
