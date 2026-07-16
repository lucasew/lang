package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.SentenceWhitespaceRuleTest

func TestSentenceWhitespaceRule_Match(t *testing.T) {
	rule := NewSentenceWhitespaceRule(map[string]string{
		"addSpaceBetweenSentences":    "Add a space between sentences",
		"missing_space_between_sentences": "Missing space between sentences",
	})

	assertGood := func(text string) {
		t.Helper()
		m := rule.MatchList(languagetool.SplitAndAnalyze(text))
		require.Equal(t, 0, len(m), "assertGood %q", text)
	}
	assertBad := func(text string) {
		t.Helper()
		m := rule.MatchList(languagetool.SplitAndAnalyze(text))
		require.Equal(t, 1, len(m), "assertBad %q", text)
	}

	assertGood("This is a text. And there's the next sentence.")
	assertGood("This is a text! And there's the next sentence.")
	assertGood("This is a text\nAnd there's the next sentence.")
	assertGood("This is a text\n\nAnd there's the next sentence.")

	assertBad("This is a text.And there's the next sentence.")
	assertBad("This is a text!And there's the next sentence.")
	assertBad("This is a text?And there's the next sentence.")
}
