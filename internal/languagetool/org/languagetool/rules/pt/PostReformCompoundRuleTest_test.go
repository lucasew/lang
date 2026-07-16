package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PostReformCompoundRuleTest.java
// Surface AnalyzePlain path (not full JLanguageTool picky mode).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPostReformCompoundRule_PostReformCompounds(t *testing.T) {
	rule := NewPostReformPortugueseCompoundRule(nil)
	require.Equal(t, "PT_COMPOUNDS_POST_REFORM", rule.GetID())

	// Hyphenated form OK
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("super-herói"))))

	// Space form flagged
	matches := rule.Match(languagetool.AnalyzePlain("super herói"))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements()[0], "super-herói")

	matches = rule.Match(languagetool.AnalyzePlain("Grã Bretanha"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Grã-Bretanha", matches[0].GetSuggestedReplacements()[0])
}
