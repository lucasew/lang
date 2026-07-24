package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.WhitespaceBeforePunctuationRuleTest

func TestWhitespaceBeforePunctuationRule_Okay(t *testing.T) {
	rule := NewWhitespaceBeforePunctuationRule(nil)
	sentence := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(" ", nil, nil), 0),
		func() *languagetool.AnalyzedTokenReadings {
			sym := "SYM"
			return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("%", &sym, nil), 1)
		}(),
	})
	require.Equal(t, 0, len(rule.Match(sentence)))
}

func TestWhitespaceBeforePunctuationRule_Error(t *testing.T) {
	rule := NewWhitespaceBeforePunctuationRule(nil)
	sentence := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("2", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(" ", nil, nil), 1),
		func() *languagetool.AnalyzedTokenReadings {
			sym := "SYM"
			return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("%", &sym, nil), 2)
		}(),
	})
	require.Equal(t, 1, len(rule.Match(sentence)))
}
