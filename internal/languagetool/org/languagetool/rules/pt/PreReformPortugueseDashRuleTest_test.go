package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PreReformPortugueseDashRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPreReformPortugueseDashRule_Test(t *testing.T) {
	rule := NewPreReformPortugueseDashRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("abaixa-língua"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("abaixa—língua"))))
}
