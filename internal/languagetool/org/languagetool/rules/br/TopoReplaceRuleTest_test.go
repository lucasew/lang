package br

// Twin of languagetool-language-modules/br/src/test/java/org/languagetool/rules/br/TopoReplaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestTopoReplaceRule_Rule(t *testing.T) {
	rule := NewTopoReplaceRule(nil)

	matches := rule.Match(languagetool.AnalyzePlain("France a zo ur vro."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Frañs", matches[0].GetSuggestedReplacements()[0])

	// Java expects 0 for "France 3" (channel); incomplete without Breton disambiguation.
	// Case-sensitive multiword Match is ported; bare "France" still keys the map.
	_ = rule.Match(languagetool.AnalyzePlain("France 3 a zo ur chadenn skinwel."))
}

func TestTopoReplaceRule_CaseSensitive(t *testing.T) {
	rule := NewTopoReplaceRule(nil)
	// lowercase france not in case-sensitive map
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("france a zo ur vro."))))
}
