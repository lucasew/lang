package sr

// Twin of languagetool-language-modules/sr/src/test/java/org/languagetool/rules/sr/SimpleGrammarEkavianReplaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sr/ekavian"
	"github.com/stretchr/testify/require"
)

func TestSimpleGrammarEkavianReplaceRule_Rule(t *testing.T) {
	rule := ekavian.NewSimpleGrammarEkavianReplaceRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Данас је диван дан."))))

	matches := rule.Match(languagetool.AnalyzePlain("Син је вишљи од оца."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"виши"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(languagetool.AnalyzePlain("У то оправдано сумљам."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"сумњам"}, matches[0].GetSuggestedReplacements())
}
