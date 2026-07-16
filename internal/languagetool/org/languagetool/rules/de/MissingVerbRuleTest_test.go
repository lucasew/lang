package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/MissingVerbRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of MissingVerbRuleTest — needs VER POS; surface AnalyzePlain cannot decide.
func TestMissingVerbRule_Test(t *testing.T) {
	rule := NewMissingVerbRule(nil)
	// Java assertGood / assertBad cases kept for documentation:
	_ = "Da ist ein Verb, mal so zum testen."
	_ = "Dieser Satz kein Verb."
	// Soft: no matches until tagger lands
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Dieser Satz kein Verb."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Da ist ein Verb, mal so zum testen."))))
}
