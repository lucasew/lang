package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugalPortugueseReplaceRule(t *testing.T) {
	rule := NewPortugalPortugueseReplaceRule(nil)

	// Example from Java: aeromoça → hospedeira de bordo
	matches := rule.Match(languagetool.AnalyzePlain("A aeromoça serviu o café."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "hospedeira de bordo", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Comprei um telefone celular."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "telemóvel", matches[0].GetSuggestedReplacements()[0])
}
