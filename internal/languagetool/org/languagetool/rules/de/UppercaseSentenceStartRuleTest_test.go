package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/UppercaseSentenceStartRuleTest.java
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/UppercaseSentenceStartRuleTest.java :: UppercaseSentenceStartRuleTest.testRule
func TestUppercaseSentenceStartRule_Rule(t *testing.T) {
	r := NewUppercaseSentenceStartRule(map[string]string{
		"incorrect_case": "This sentence does not start with an uppercase letter",
	})
	analyze := func(s string) []*languagetool.AnalyzedSentence {
		if strings.Contains(s, ". ") || strings.Contains(s, "! ") || strings.Contains(s, "? ") {
			return languagetool.SplitAndAnalyze(s)
		}
		return []*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}
	}
	// good
	require.Empty(t, r.MatchList(analyze("Dies ist ein Satz. Und hier kommt noch einer")))
	require.Empty(t, r.MatchList(analyze("Dies ist ein Satz. Ätsch, noch einer mit Umlaut.")))
	require.Empty(t, r.MatchList(analyze("\"Dies ist ein Satz!\"")))
	// bad
	require.Equal(t, 2, len(r.MatchList(analyze("etwas beginnen. und der auch nicht"))))
	require.Equal(t, 1, len(r.MatchList(analyze("schön!"))))
	require.Equal(t, 1, len(r.MatchList(analyze("Dies ist ein Satz. ätsch, noch einer mit Umlaut."))))
	require.Equal(t, 1, len(r.MatchList(analyze("\"dies ist ein Satz!\""))))
}
