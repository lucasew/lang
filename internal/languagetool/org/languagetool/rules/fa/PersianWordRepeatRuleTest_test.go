package fa

// Outcome twins for org.languagetool.rules.fa.PersianWordRepeatRule.
// Java has no dedicated *Test.java; phrases come from Java ignore() comments + example pair.

import (
	"testing"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func persianWordRepeatMessages() map[string]string {
	return map[string]string{
		"repetition":            "تکرار",
		"desc_repetition":       "تکرار کلمه",
		"desc_repetition_short": "تکرار",
	}
}

func matchCountFA(t *testing.T, rule *PersianWordRepeatRule, input string) int {
	t.Helper()
	return len(rule.Match(languagetool.AnalyzePlain(input)))
}

// ignoreAtFA mirrors PortugueseWordRepeatRuleTest.ignore on AnalyzePlain tokens.
func ignoreAtFA(t *testing.T, rule *PersianWordRepeatRule, input string, position int) bool {
	t.Helper()
	tokens := languagetool.AnalyzePlain(input).GetTokensWithoutWhitespace()
	return rule.Ignore(tokens, position)
}

// utf16SubstrFA slices text by Java UTF-16 code-unit offsets (token startPos/toPos).
func utf16SubstrFA(s string, from, to int) string {
	u := utf16.Encode([]rune(s))
	if from < 0 {
		from = 0
	}
	if to > len(u) {
		to = len(u)
	}
	if from >= to {
		return ""
	}
	return string(utf16.Decode(u[from:to]))
}

// Java example pair + ignore-comment phrases (0 matches) and one clear false-alarm (1 match).
func TestPersianWordRepeatRule_Rule(t *testing.T) {
	rule := NewPersianWordRepeatRule(persianWordRepeatMessages())

	// Example (Java addExamplePair wrong sentence without markers)
	require.Equal(t, 1, matchCountFA(t, rule, "این کار برای برای تو بود."))

	// FA ignore list — phrases from Java PersianWordRepeatRule.ignore comments
	require.Equal(t, 0, matchCountFA(t, rule, "من لی لی را دیدم"))
	require.Equal(t, 0, matchCountFA(t, rule, "آب درون بطری ۱۰۰ سی سی بود"))
	require.Equal(t, 0, matchCountFA(t, rule, "لک لک یک پرنده است"))
	require.Equal(t, 0, matchCountFA(t, rule, "غذایش را ریز ریز کرد"))
	require.Equal(t, 0, matchCountFA(t, rule, "جز جز این کتاب را بلدم"))
	require.Equal(t, 0, matchCountFA(t, rule, "با من کل کل نکن"))

	// Real repeats that are not on the FA list must still fire
	require.Equal(t, 1, matchCountFA(t, rule, "من من اینجا هستم"))
}

// ignore() FA arms + super.ignore name list after ExtraIgnore returns false.
// Positions are the second token of the repeated pair (AnalyzePlain tokensWithoutWhitespace).
func TestPersianWordRepeatRule_Ignore(t *testing.T) {
	rule := NewPersianWordRepeatRule(persianWordRepeatMessages())

	// wordRepetitionOf exact equals at the second of the repeated pair
	// SENT_START=0: من=1 لی=2 لی=3
	require.True(t, ignoreAtFA(t, rule, "من لی لی را دیدم", 3))
	// آب=1 درون=2 بطری=3 ۱۰۰=4 سی=5 سی=6
	require.True(t, ignoreAtFA(t, rule, "آب درون بطری ۱۰۰ سی سی بود", 6))
	// لک=1 لک=2
	require.True(t, ignoreAtFA(t, rule, "لک لک یک پرنده است", 2))
	// غذایش=1 را=2 ریز=3 ریز=4
	require.True(t, ignoreAtFA(t, rule, "غذایش را ریز ریز کرد", 4))
	// جز=1 جز=2
	require.True(t, ignoreAtFA(t, rule, "جز جز این کتاب را بلدم", 2))
	// با=1 من=2 کل=3 کل=4
	require.True(t, ignoreAtFA(t, rule, "با من کل کل نکن", 4))

	// Non-listed repeat is not ignored (این=1 کار=2 برای=3 برای=4)
	require.False(t, ignoreAtFA(t, rule, "این کار برای برای تو بود.", 4))

	// super.ignore names (Java return super.ignore after FA checks)
	require.True(t, ignoreAtFA(t, rule, "Phi Phi Islands", 2))
	require.True(t, ignoreAtFA(t, rule, "Li Li", 2))
}

// wordRepetitionOf uses Token.equals (case-sensitive), not EqualFold.
// Persian script has no case, but Latin super names and the helper stay exact.
func TestPersianWordRepeatRule_WordRepetitionOfCaseSensitive(t *testing.T) {
	tokens := languagetool.AnalyzePlain("لی لی").GetTokensWithoutWhitespace()
	// tokens: SENT_START, لی, لی — position of second لی is 2
	require.True(t, wordRepetitionOfFA("لی", tokens, 2))
	require.False(t, wordRepetitionOfFA("لی", tokens, 1))
	require.False(t, wordRepetitionOfFA("سی", tokens, 2))

	// Latin super list is case-sensitive via base wordRepetitionOf
	rule := NewPersianWordRepeatRule(nil)
	require.True(t, ignoreAtFA(t, rule, "Li Li", 2))
	require.False(t, ignoreAtFA(t, rule, "li li", 2))
	require.False(t, ignoreAtFA(t, rule, "LI LI", 2))
}

// createRuleMatch outcomes: id, message, short message, span, suggestion.
func TestPersianWordRepeatRule_MatchSpanAndSuggestion(t *testing.T) {
	msgs := persianWordRepeatMessages()
	rule := NewPersianWordRepeatRule(msgs)
	text := "این کار برای برای تو بود."
	sent := languagetool.AnalyzePlain(text)
	matches := rule.Match(sent)
	require.Len(t, matches, 1)
	m := matches[0]
	require.Equal(t, "PERSIAN_WORD_REPEAT_RULE", m.GetSpecificRuleId())
	require.Equal(t, msgs["repetition"], m.Message)
	require.Equal(t, msgs["desc_repetition_short"], m.ShortMessage)
	require.Equal(t, []string{"برای"}, m.GetSuggestedReplacements())
	// Java: fromPos of first برای through pos+prevToken.length() of second (UTF-16)
	require.Equal(t, "برای برای", utf16SubstrFA(text, m.FromPos, m.ToPos))
}

// Java getId + example pair + estimateContextForSureMatch + description key.
func TestPersianWordRepeatRule_Metadata(t *testing.T) {
	msgs := persianWordRepeatMessages()
	rule := NewPersianWordRepeatRule(msgs)
	require.Equal(t, "PERSIAN_WORD_REPEAT_RULE", rule.GetID())
	require.Equal(t, 1, rule.EstimateContextForSureMatch())
	require.Equal(t, msgs["desc_repetition"], rule.GetDescription())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "این کار <marker>برای برای</marker> تو بود.", inc[0].GetExample())
	require.Equal(t, []string{"برای"}, inc[0].GetCorrections())
	require.Equal(t, "این کار <marker>برای</marker> تو بود.", rule.GetCorrectExamples()[0].GetExample())
}
