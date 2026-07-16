package sr

// Twin of languagetool-language-modules/sr/src/test/java/org/languagetool/rules/sr/SimpleStyleEkavianReplaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sr/ekavian"
	"github.com/stretchr/testify/require"
)

func TestSimpleStyleEkavianReplaceRule_Rule(t *testing.T) {
	rule := ekavian.NewSimpleStyleEkavianReplaceRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Он је добар."))))

	matches := rule.Match(languagetool.AnalyzePlain("Она је дебела."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"елегантно попуњена"}, matches[0].GetSuggestedReplacements())
}
