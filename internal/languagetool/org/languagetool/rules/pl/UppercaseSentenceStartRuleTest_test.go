package pl

// Twin of UppercaseSentenceStartRuleTest (Polish)
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of UppercaseSentenceStartRuleTest.testPolishSpecialCases (subset)
func TestUppercaseSentenceStartRule_PolishSpecialCases(t *testing.T) {
	r := NewUppercaseSentenceStartRule(map[string]string{
		"incorrect_case": "Zdanie nie zaczyna się wielką literą",
	})
	analyze := func(s string) []*languagetool.AnalyzedSentence {
		if strings.Contains(s, ". ") {
			return languagetool.SplitAndAnalyze(s)
		}
		return []*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}
	}
	require.Empty(t, r.MatchList(analyze("To jest zdanie.")))
	require.Equal(t, 1, len(r.MatchList(analyze("to jest zdanie."))))
	// second sentence lowercase
	require.Equal(t, 1, len(r.MatchList(analyze("Hello. world is small."))))
}
