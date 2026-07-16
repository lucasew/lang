package ga

// Twin of languagetool-language-modules/ga/src/test/java/org/languagetool/rules/ga/DhaNoBeirtRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDhaNoBeirtRule_Rule(t *testing.T) {
	rule := NewDhaNoBeirtRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Seo abairt bheag."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Tá beirt dheartháireacha agam."))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Tá dhá dheartháireacha agam."))))
	require.Equal(t, 2, len(rule.Match(languagetool.AnalyzePlain("Seo dhá ab déag"))))
	require.Equal(t, 2, len(rule.Match(languagetool.AnalyzePlain("Tá dhá dheartháireacha níos aosta déag agam."))))
}
