package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/UppercaseSentenceStartRuleTest.java
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Port of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/UppercaseSentenceStartRuleTest.java :: UppercaseSentenceStartRuleTest.testRule
func TestUppercaseSentenceStartRule_Rule(t *testing.T) {
	r := rules.NewUppercaseSentenceStartRule(map[string]string{
		"incorrect_case": "This sentence does not start with an uppercase letter",
	}, "en")
	analyze := func(s string) []*languagetool.AnalyzedSentence {
		if strings.Contains(s, ". ") || strings.Contains(s, "! ") {
			return languagetool.SplitAndAnalyze(s)
		}
		return []*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}
	}
	require.Empty(t, r.MatchList(analyze("This is a test.")))
	require.Empty(t, r.MatchList(analyze("http://example.com")))
	require.Equal(t, 1, len(r.MatchList(analyze("this is a test sentence."))))
	require.Equal(t, 1, len(r.MatchList(analyze("this!"))))
	// second sentence lowercase
	require.Equal(t, 1, len(r.MatchList(analyze("Hello. world is small."))))
}
