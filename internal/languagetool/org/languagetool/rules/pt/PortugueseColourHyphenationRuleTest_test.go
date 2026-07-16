package pt

// Twin of PortugueseColourHyphenationRule — surface compound check for colour names.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseColourHyphenationRule(t *testing.T) {
	rule := NewPortugueseColourHyphenationRule(nil)
	require.Equal(t, "PT_COLOUR_HYPHENATION", rule.GetID())
	// "amarelo claro" should suggest hyphenated form (data uses * = hyphen only).
	matches := rule.Match(languagetool.AnalyzePlain("uma parede amarelo claro"))
	require.NotEmpty(t, matches)
	require.Contains(t, matches[0].GetSuggestedReplacements()[0], "amarelo-claro")
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("uma parede amarelo-claro"))))
}
