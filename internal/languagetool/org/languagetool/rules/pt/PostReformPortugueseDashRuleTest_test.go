package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PostReformPortugueseDashRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPostReformPortugueseDashRule_Test(t *testing.T) {
	rule := NewPostReformPortugueseDashRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("ab-reação"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("ab—reação"))))
}
