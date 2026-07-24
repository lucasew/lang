package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.WhitespaceBeforePunctuationRuleTest — Java-visible outcomes only.
//
// Java setup:
//
//	rule = new WhitespaceBeforePunctuationRule(TestTools.getEnglishMessages());
//	AnalyzedSentence built from hand-made AnalyzedTokenReadings[]
//	assertThat(rule.match(sentence).length, is(0|1));

func engMessagesWhitespaceBeforePunctuation() map[string]string {
	// MessagesBundle_en keys used by WhitespaceBeforePunctuationRule.
	return map[string]string{
		"desc_whitespace_before_punctuation": "Use of whitespace before colon, semicolon and percentage.",
		"no_space_before_colon":              "Don't put a space before the colon.",
		"no_space_before_semicolon":          "Don't put a space before the semicolon.",
		"no_space_before_percentage":         "Don't put a space before the percentage.",
	}
}

// TestWhitespaceBeforePunctuationRule_Okay ports testOkay:
// tokens [" ", "%"/SYM] → match length 0.
func TestWhitespaceBeforePunctuationRule_Okay(t *testing.T) {
	rule := NewWhitespaceBeforePunctuationRule(engMessagesWhitespaceBeforePunctuation())
	// Java:
	// new AnalyzedSentence(new AnalyzedTokenReadings[] {
	//   new AnalyzedTokenReadings(new AnalyzedToken(" ", null, null), 0),
	//   new AnalyzedTokenReadings(new AnalyzedToken("%", "SYM", null), 1)
	// });
	sym := "SYM"
	sentence := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(" ", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("%", &sym, nil), 1),
	})
	require.Equal(t, 0, len(rule.Match(sentence)))
}

// TestWhitespaceBeforePunctuationRule_Error ports testError:
// tokens ["2", " ", "%"/SYM] → match length 1.
func TestWhitespaceBeforePunctuationRule_Error(t *testing.T) {
	rule := NewWhitespaceBeforePunctuationRule(engMessagesWhitespaceBeforePunctuation())
	// Java:
	// new AnalyzedSentence(new AnalyzedTokenReadings[] {
	//   new AnalyzedTokenReadings(new AnalyzedToken("2", null, null), 0),
	//   new AnalyzedTokenReadings(new AnalyzedToken(" ", null, null), 1),
	//   new AnalyzedTokenReadings(new AnalyzedToken("%", "SYM", null), 2)
	// });
	sym := "SYM"
	sentence := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("2", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(" ", nil, nil), 1),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("%", &sym, nil), 2),
	})
	require.Equal(t, 1, len(rule.Match(sentence)))
}
