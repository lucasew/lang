package pt

// Unit tests for Portuguese Wikipedia / Redundancy / Wordiness ASR2 ports.
// No dedicated Java Wikipedia/Redundancy/Wordiness twin tests in the PT module;
// assertions follow rule examples and dictionary entries.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseWikipediaRule(t *testing.T) {
	rule := NewPortugueseWikipediaRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("a diversos pedidos."))))

	matches := rule.Match(languagetool.AnalyzePlain("Respondeu à diversos pedidos."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "a diversos", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Ao meu ver, está errado."))
	require.Equal(t, 1, len(matches))
	// Sentence-start capitalization of multiword suggestion
	require.Equal(t, "A meu ver", matches[0].GetSuggestedReplacements()[0])
}

func TestPortugueseRedundancyRule(t *testing.T) {
	rule := NewPortugueseRedundancyRule(nil)
	// Example from Java rule: duna de areia → duna
	matches := rule.Match(languagetool.AnalyzePlain("A duna de areia era enorme."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "duna", matches[0].GetSuggestedReplacements()[0])

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("A duna era enorme."))))
}

func TestPortugueseWordinessRule(t *testing.T) {
	rule := NewPortugueseWordinessRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("A forma como ele fala impressiona."))
	require.Equal(t, 1, len(matches))
	// First token of match is capitalized → suggestion capitalised
	require.Equal(t, "Como", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Vimos a forma como ele fala."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "como", matches[0].GetSuggestedReplacements()[0])

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Como ele fala impressiona."))))
}
