package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishUnpairedQuotesRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishUnpairedQuotesRule_Rule(t *testing.T) {
	rule := NewEnglishUnpairedQuotesRule(nil)
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	require.Equal(t, 0, matchN("This is a word 'test'."))
	require.Equal(t, 0, matchN("This is what he said: \"We believe in freedom. This is what we do.\""))
	// Unpaired double quotes; inject POS on contraction apostrophe (Java disambig/tagger).
	// Without POS, EN override treats ' as a quote mark (fail closed — no invent).
	sent := languagetool.AnalyzePlain("\"I'm over here, she said.")
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "'" {
			pos := "_apostrophe_contraction_"
			tok.AddReading(languagetool.NewAnalyzedToken("'", &pos, nil), "test")
		}
	}
	require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{sent})))
}

func TestEnglishUnpairedQuotesRule_FailClosedWithoutPOS(t *testing.T) {
	rule := NewEnglishUnpairedQuotesRule(nil)
	// Apostrophe in I'm is not exempt without POS tags (Java POS-gated override).
	n := len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("\"I'm over here, she said.")}))
	require.Equal(t, 2, n)
}
