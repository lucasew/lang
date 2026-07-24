package pt

// Twin of UppercaseSentenceStartRuleTest (Portuguese)
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Port of UppercaseSentenceStartRuleTest.testUppercaseRule
func TestUppercaseSentenceStartRule_UppercaseRule(t *testing.T) {
	r := rules.NewUppercaseSentenceStartRule(map[string]string{
		"incorrect_case": "A frase não começa com maiúscula",
	}, "pt")
	analyze := func(s string) []*languagetool.AnalyzedSentence {
		if strings.Contains(s, ". ") {
			return languagetool.SplitAndAnalyze(s)
		}
		return []*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}
	}
	require.Empty(t, r.MatchList(analyze("Isto é uma frase.")))
	require.Equal(t, 1, len(r.MatchList(analyze("isto é uma frase."))))
	require.Equal(t, 1, len(r.MatchList(analyze("Olá. mundo é grande."))))
}
