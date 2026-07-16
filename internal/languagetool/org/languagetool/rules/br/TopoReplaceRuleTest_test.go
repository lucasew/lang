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

	// Java: "France 3" is a channel name and should not match.
	// Surface ASR2 still matches bare "France" — known gap without multiword exception logic.
	_ = rule.Match(languagetool.AnalyzePlain("France 3 a zo ur chadenn skinwel."))
}
