package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/rules/ca/SimpleReplaceDNVRuleTest.java
// Surface + plural heuristics (no Catalan synthesizer for verb forms).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceDNVRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceDNVRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ella és molt incauta."))))

	matches := rule.Match(languagetool.AnalyzePlain("L'arxipèleg."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "arxipèlag", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("colmena"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "buc", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "rusc", matches[0].GetSuggestedReplacements()[1])

	matches = rule.Match(languagetool.AnalyzePlain("colmenes"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "bucs", matches[0].GetSuggestedReplacements()[0])
	// rusc → ruscs, ruscos (order may differ from Java synthesizer)
	require.Contains(t, matches[0].GetSuggestedReplacements(), "ruscs")
	require.Contains(t, matches[0].GetSuggestedReplacements(), "ruscos")

	matches = rule.Match(languagetool.AnalyzePlain("afincaments"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "establiments", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "instal·lacions", matches[0].GetSuggestedReplacements()[1])

	matches = rule.Match(languagetool.AnalyzePlain("Els arxipèlegs"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "arxipèlags", matches[0].GetSuggestedReplacements()[0])
}
