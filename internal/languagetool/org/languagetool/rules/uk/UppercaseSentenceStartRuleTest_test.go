package uk

// Twin of UppercaseSentenceStartRuleTest (Ukrainian)
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Port of UppercaseSentenceStartRuleTest.testUkrainian
func TestUppercaseSentenceStartRule_Ukrainian(t *testing.T) {
	r := rules.NewUppercaseSentenceStartRule(map[string]string{
		"incorrect_case": "Речення не починається з великої літери",
	}, "uk")
	analyze := func(s string) []*languagetool.AnalyzedSentence {
		if strings.Contains(s, ". ") {
			return languagetool.SplitAndAnalyze(s)
		}
		return []*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}
	}
	require.Empty(t, r.MatchList(analyze("Це речення.")))
	require.Equal(t, 1, len(r.MatchList(analyze("це речення."))))
	// second sentence
	require.Equal(t, 1, len(r.MatchList(analyze("Привіт. світ."))))
}
