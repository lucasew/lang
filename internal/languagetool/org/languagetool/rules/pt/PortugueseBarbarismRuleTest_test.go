package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PortugueseBarbarismRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseBarbarismRule_ReplaceBarbarisms(t *testing.T) {
	rule := NewPortugueseBarbarismsRule(nil)

	// Exceptions (named entities / multi-token phrases) — no matches in pt-BR list.
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("New York Stock Exchange"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Yankee Doodle, faça o morra"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("mas inferior ao Opera Browser."))))

	// Positive: dictionary entry from pt-BR barbarisms (app → aplicativo)
	matches := rule.Match(languagetool.AnalyzePlain("Baixei um app novo."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "aplicativo", matches[0].GetSuggestedReplacements()[0])
}
