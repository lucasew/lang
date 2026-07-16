package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/rules/ca/SimpleReplaceDNVSecondaryRuleTest.java
// Surface port: dispost stand-in for lemma dispondre (no synthesizer/tagger).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceDNVSecondaryRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceDNVSecondaryRule(nil)

	// incorrect
	matches := rule.Match(languagetool.AnalyzePlain("S'ha dispost a fer-ho."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "disposat", matches[0].GetSuggestedReplacements()[0])

	// Surface exact lemma from secondary list
	matches = rule.Match(languagetool.AnalyzePlain("Un armatost enorme."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "baluerna", matches[0].GetSuggestedReplacements()[0])
}
