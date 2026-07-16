package gl

// Unit tests for Galician dictionary-backed rule ports (examples from Java rule classes).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGalicianWikipediaRule(t *testing.T) {
	rule := NewGalicianWikipediaRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Útil a efectos de control."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "para os efectos de", matches[0].GetSuggestedReplacements()[0])
}

func TestGalicianRedundancyRule(t *testing.T) {
	rule := NewGalicianRedundancyRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("A duna de area era grande."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "duna", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("A duna era grande."))))
}

func TestGalicianWordinessRule(t *testing.T) {
	rule := NewGalicianWordinessRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Raramente é o caso en que acontece isto."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Raramente acontece", matches[0].GetSuggestedReplacements()[0])
}

func TestGalicianBarbarismsRule(t *testing.T) {
	rule := NewGalicianBarbarismsRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Envíe o curriculum vitae."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "currículo", matches[0].GetSuggestedReplacements()[0])
}

func TestSimpleReplaceRule(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Alomenos tres persoas."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Polo menos", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Hai alomenos tres persoas."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "polo menos", matches[0].GetSuggestedReplacements()[0])
}

func TestCastWordsRule(t *testing.T) {
	rule := NewCastWordsRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Camiñaba pola acera."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "beirarrúa", matches[0].GetSuggestedReplacements()[0])
}
