package en

import (
	"fmt"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Java en.LongSentenceRule: setUrl + getShortMessage; base STYLE / picky / ID from core.
func TestLongSentenceRule_Metadata(t *testing.T) {
	// Java: new LongSentenceRule(messages, new UserConfig(), 40)
	rule := NewLongSentenceRule(nil, languagetool.NewUserConfig(), 40)
	require.Equal(t, "TOO_LONG_SENTENCE", rule.GetID())
	require.Equal(t, "Long sentence", rule.GetShortMessage())
	require.Equal(t, "Long sentence", rule.ShortMsg)
	require.Contains(t, rule.GetURL(), "splitting-long-sentences")
	require.Equal(t, "https://languagetool.org/insights/post/splitting-long-sentences/", rule.GetURL())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	require.True(t, rule.HasTag(rules.TagPicky))
	require.Equal(t, 40, rule.MaxWords)
	require.Equal(t, 0, rule.MinToCheckParagraph())
}

// UserConfig.getConfigValueByID(getId()) overrides defaultWords — same path as core Java.
func TestLongSentenceRule_UserConfigMaxWords(t *testing.T) {
	uc := languagetool.NewUserConfig()
	uc.SetConfigValueByID(rules.LongSentenceRuleID, []any{10})
	rule := NewLongSentenceRule(nil, uc, 40)
	require.Equal(t, 10, rule.MaxWords)
}

// Match / no-match via base MatchList with defaultWords threshold (EN subclass still detects).
// Spans mirror Java LongSentenceRuleTest shortRule path (maxWords=6) with AnalyzeTextLocal.
func TestLongSentenceRule_Match(t *testing.T) {
	msgs := map[string]string{
		"long_sentence_rule_msg2": "This sentence is over {0} words long. Consider splitting it up, as shorter sentences make the text easier to read.",
	}
	// Java shortRule: new LongSentenceRule(..., new UserConfig(), 6)
	rule := NewLongSentenceRule(msgs, languagetool.NewUserConfig(), 6)
	require.Equal(t, "TOO_LONG_SENTENCE", rule.GetID())
	require.Equal(t, "Long sentence", rule.GetShortMessage())
	require.Contains(t, rule.GetURL(), "splitting-long-sentences")

	assertNoMatch := func(input string) {
		t.Helper()
		ms := rule.MatchList(languagetool.AnalyzeTextLocal(input))
		require.Equal(t, 0, len(ms), "assertNoMatch %q got %v", input, enLSRSpans(ms))
	}
	assertMatch := func(input string, from, to int) {
		t.Helper()
		ms := rule.MatchList(languagetool.AnalyzeTextLocal(input))
		require.Equal(t, 1, len(ms), "assertMatch %q got %d spans %v", input, len(ms), enLSRSpans(ms))
		require.Equal(t, from, ms[0].GetFromPos(), "fromPos for %q", input)
		require.Equal(t, to, ms[0].GetToPos(), "toPos for %q", input)
	}

	// short / under threshold
	assertNoMatch("This is a rather short text.")
	assertNoMatch("one two three four five six.")
	// at / over threshold
	assertMatch("This is also a rather short text.", 0, 33)
	assertMatch("one two three four five six seven.", 0, 34)
	// non-word tokens do not count
	assertNoMatch("These ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ don't count.")
}

func enLSRSpans(ms []*rules.RuleMatch) string {
	if len(ms) == 0 {
		return "[]"
	}
	s := "["
	for i, m := range ms {
		if i > 0 {
			s += " "
		}
		s += fmt.Sprintf("%d-%d", m.GetFromPos(), m.GetToPos())
	}
	return s + "]"
}
