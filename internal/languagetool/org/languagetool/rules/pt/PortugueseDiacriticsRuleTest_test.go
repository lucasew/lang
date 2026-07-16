package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PortugueseDiacriticsRuleTest.java
//
// The Java test exercises grammar rulegroup DIACRITICS (dialect bebé/bebê via full JLanguageTool).
// This twin covers PortugueseDiacriticsRule (ASR2 /pt/diacritics.txt) directly.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseDiacriticsRule_Test(t *testing.T) {
	rule := NewPortugueseDiacriticsRule(nil)

	// Example from Java rule: coupe → coupé
	matches := rule.Match(languagetool.AnalyzePlain("Um coupe clássico."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "coupé", matches[0].GetSuggestedReplacements()[0])

	// Multiword: a la → à la
	matches = rule.Match(languagetool.AnalyzePlain("Servido a la carte."))
	require.GreaterOrEqual(t, len(matches), 1)
	require.Contains(t, matches[0].GetSuggestedReplacements()[0], "à la")

	// Correct form should not match
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Um coupé clássico."))))
}
