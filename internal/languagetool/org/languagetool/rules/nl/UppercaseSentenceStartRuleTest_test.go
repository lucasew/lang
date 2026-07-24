package nl

// Twin of UppercaseSentenceStartRuleTest (Dutch)
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Port of UppercaseSentenceStartRuleTest.testDutchSpecialCases (subset)
func TestUppercaseSentenceStartRule_DutchSpecialCases(t *testing.T) {
	r := rules.NewUppercaseSentenceStartRule(map[string]string{
		"incorrect_case": "Zin begint niet met hoofdletter",
	}, "nl")
	analyze := func(s string) []*languagetool.AnalyzedSentence {
		if strings.Contains(s, ". ") {
			return languagetool.SplitAndAnalyze(s)
		}
		return []*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}
	}
	require.Empty(t, r.MatchList(analyze("Dit is een zin.")))
	require.Equal(t, 1, len(r.MatchList(analyze("dit is een zin."))))
}
