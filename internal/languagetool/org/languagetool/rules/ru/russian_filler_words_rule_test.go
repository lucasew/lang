package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestRussianFillerWordsRule_IDAndDefaultOff(t *testing.T) {
	rule := NewRussianFillerWordsRule(nil)
	require.Equal(t, "FILLER_WORDS_RU", rule.GetID())
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, rules.FillerWordsDefaultMinPercent, rule.GetMinPercent())
	require.True(t, rule.ExcludeDirectSpeech)
}

func TestRussianFillerWordsRule_FillerSetSize(t *testing.T) {
	require.Equal(t, 15, len(russianFillerWords))
}

func TestRussianFillerWordsRule_IsFillerWordCaseSensitive(t *testing.T) {
	rule := NewRussianFillerWordsRule(nil)
	require.True(t, rule.isFillerWord("эээ"))
	require.True(t, rule.isFillerWord("э"))
	require.True(t, rule.isFillerWord("однако"))
	require.True(t, rule.isFillerWord("ух-ты"))
	// Java HashSet.contains is case-sensitive — capitalised forms are not fillers.
	require.False(t, rule.isFillerWord("Эээ"))
	require.False(t, rule.isFillerWord("Э"))
	require.False(t, rule.isFillerWord("Однако"))
	require.False(t, rule.isFillerWord("not_a_filler"))
}

// matchedSurface uses rune indices — AnalyzePlain positions match Java String offsets (BMP).
func matchedSurface(sentence *languagetool.AnalyzedSentence, m *rules.RuleMatch) string {
	text := sentence.GetText()
	from, to := m.GetFromPos(), m.GetToPos()
	runes := []rune(text)
	if from < 0 || to > len(runes) || from > to {
		return ""
	}
	return string(runes[from:to])
}

func TestRussianFillerWordsRule_IsExceptionAlwaysFalse(t *testing.T) {
	// Unlike PT "mas" after comma, RU isException is always false.
	// "однако" after comma is still a filler hit (minPercent 0).
	rule := NewRussianFillerWordsRuleWithMinPercent(nil, 0)
	sent := languagetool.AnalyzePlain("X, однако Z.")
	matches := rule.Match(sent)
	require.Len(t, matches, 1)
	require.Equal(t, "однако", matchedSurface(sent, matches[0]))
}

func TestRussianFillerWordsRule_MinPercent0ShowsFillers(t *testing.T) {
	rule := NewRussianFillerWordsRuleWithMinPercent(nil, 0)
	// Many non-filler words, one filler "эээ" — still shown at 0%.
	// Avoid set members: ах аа ааа аааа ау бу вау ох однако эээ э эй эх ух-ты ух
	long := "Он пришел домой после работы и встретил друзей вечером эээ потом ушел рано."
	sent := languagetool.AnalyzePlain(long)
	matches := rule.Match(sent)
	require.Len(t, matches, 1)
	require.Equal(t, "эээ", matchedSurface(sent, matches[0]))
}

func TestRussianFillerWordsRule_Default8PercentThreshold(t *testing.T) {
	rule := NewRussianFillerWordsRule(nil)
	require.Equal(t, 8, rule.GetMinPercent())

	// Short sentence: 1 filler / few words → over 8%.
	// "Он эээ пришел." ≈ 3 words, 1 filler → ~33% > 8
	sentShort := languagetool.AnalyzePlain("Он эээ пришел.")
	matches := rule.Match(sentShort)
	require.Len(t, matches, 1)
	require.Equal(t, "эээ", matchedSurface(sentShort, matches[0]))

	// Long sentence: 1 filler under 8% → no match.
	// Need wordCount > 12.5 for 1/n <= 8% → n >= 13 words.
	long := "Он пришел домой после работы и встретил друзей вечером эээ потом ушел рано."
	matches = rule.Match(languagetool.AnalyzePlain(long))
	require.Equal(t, 0, len(matches))
}

func TestRussianFillerWordsRule_CaseSensitiveMatch(t *testing.T) {
	rule := NewRussianFillerWordsRuleWithMinPercent(nil, 0)
	// Sentence-initial capitalised filler must not match (case-sensitive isFillerWord).
	// "Эээ" is not in the set; only "эээ".
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Эээ пришел."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Э сказал."))))
	sent := languagetool.AnalyzePlain("Он эээ пришел.")
	matches := rule.Match(sent)
	require.Len(t, matches, 1)
	require.Equal(t, "эээ", matchedSurface(sent, matches[0]))
}

func TestRussianFillerWordsRule_UserConfigMinPercent(t *testing.T) {
	uc := languagetool.NewUserConfig()
	uc.SetConfigValueByID("FILLER_WORDS_RU", []any{0})
	rule := NewRussianFillerWordsRuleWithUserConfig(nil, uc)
	require.Equal(t, 0, rule.GetMinPercent())

	long := "Он пришел домой после работы и встретил друзей вечером эээ потом ушел рано."
	sent := languagetool.AnalyzePlain(long)
	matches := rule.Match(sent)
	require.Len(t, matches, 1)
	require.Equal(t, "эээ", matchedSurface(sent, matches[0]))
}

func TestRussianFillerWordsRule_MessageKeysNotInvented(t *testing.T) {
	// Without messages, Description must not be invented RU/EN copy.
	rule := NewRussianFillerWordsRule(nil)
	require.NotEqual(t, "Filler words", rule.GetDescription())
	require.NotContains(t, rule.GetDescription(), "паразит")
	require.NotContains(t, rule.GetDescription(), "Filler word")

	msgs := map[string]string{
		"filler_words_rule_desc": "desc-from-messages",
		"filler_words_rule_msg":  "msg-from-messages",
	}
	rule2 := NewRussianFillerWordsRuleWithMinPercent(msgs, 0)
	require.Equal(t, "desc-from-messages", rule2.GetDescription())
	matches := rule2.Match(languagetool.AnalyzePlain("Он эээ пришел."))
	require.NotEmpty(t, matches)
	require.Equal(t, "msg-from-messages", matches[0].GetMessage())
}
