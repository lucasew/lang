package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestPortugueseFillerWordsRule_IDAndDefaultOff(t *testing.T) {
	rule := NewPortugueseFillerWordsRule(nil)
	require.Equal(t, "FILLER_WORDS_PT", rule.GetID())
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, rules.FillerWordsDefaultMinPercent, rule.GetMinPercent())
	require.True(t, rule.ExcludeDirectSpeech)
}

func TestPortugueseFillerWordsRule_FillerSetSize(t *testing.T) {
	require.Equal(t, 172, len(portugueseFillerWords))
}

func TestPortugueseFillerWordsRule_IsFillerWordCaseSensitive(t *testing.T) {
	rule := NewPortugueseFillerWordsRule(nil)
	require.True(t, rule.isFillerWord("realmente"))
	require.True(t, rule.isFillerWord("mas"))
	require.True(t, rule.isFillerWord("não"))
	// Java HashSet.contains is case-sensitive — capitalised forms are not fillers.
	require.False(t, rule.isFillerWord("Realmente"))
	require.False(t, rule.isFillerWord("REALMENTE"))
	require.False(t, rule.isFillerWord("Mas"))
	require.False(t, rule.isFillerWord("not_a_filler"))
}

func matchedSurface(sentence *languagetool.AnalyzedSentence, m *rules.RuleMatch) string {
	text := sentence.GetText()
	from, to := m.GetFromPos(), m.GetToPos()
	if from < 0 || to > len(text) || from > to {
		return ""
	}
	return text[from:to]
}

func TestPortugueseFillerWordsRule_MasAfterCommaException(t *testing.T) {
	// minPercent 0 → show all fillers; "mas" after comma is exception.
	rule := NewPortugueseFillerWordsRuleWithMinPercent(nil, 0)
	sent := languagetool.AnalyzePlain("X, mas Z.")
	require.Equal(t, 0, len(rule.Match(sent)))

	// "mas" without preceding comma is a filler.
	// Avoid other fillers: "X" and "Z" are not in the set.
	sent2 := languagetool.AnalyzePlain("X mas Z.")
	matches := rule.Match(sent2)
	require.Len(t, matches, 1)
	require.Equal(t, "mas", matchedSurface(sent2, matches[0]))
}

func TestPortugueseFillerWordsRule_MinPercent0ShowsFillers(t *testing.T) {
	rule := NewPortugueseFillerWordsRuleWithMinPercent(nil, 0)
	// Many non-filler words, one filler "realmente" — still shown at 0%.
	// Avoid list members: para, com is ok? "com" not in list. "e" not. "a" not. "da" not.
	// "de" not. "o" not. "do" not. "trabalho" not. "jantar" not. "casa" not.
	// "veio" not. "gostou" not. "noite" not. "juntos" not. "apos" not (sem accent; "após" not in list either as exact).
	// Careful: "mais" is filler — avoid. "muito" filler — avoid.
	long := "Ele veio de casa com amigos familia jantar gostou noite juntos apos trabalho e realmente saiu cedo ontem."
	sent := languagetool.AnalyzePlain(long)
	matches := rule.Match(sent)
	require.Len(t, matches, 1)
	require.Equal(t, "realmente", matchedSurface(sent, matches[0]))
}

func TestPortugueseFillerWordsRule_Default8PercentThreshold(t *testing.T) {
	rule := NewPortugueseFillerWordsRule(nil)
	require.Equal(t, 8, rule.GetMinPercent())

	// Short sentence: 1 filler / few words → over 8%.
	// "Ele realmente veio." ≈ 3 words, 1 filler → ~33% > 8
	sentShort := languagetool.AnalyzePlain("Ele realmente veio.")
	matches := rule.Match(sentShort)
	require.Len(t, matches, 1)
	require.Equal(t, "realmente", matchedSurface(sentShort, matches[0]))

	// Long sentence: 1 filler under 8% → no match.
	// Need wordCount > 12.5 for 1/n <= 8% → n >= 13 words.
	long := "Ele veio de casa com amigos familia jantar gostou noite juntos apos trabalho e realmente saiu cedo ontem."
	matches = rule.Match(languagetool.AnalyzePlain(long))
	require.Equal(t, 0, len(matches))
}

func TestPortugueseFillerWordsRule_CaseSensitiveMatch(t *testing.T) {
	rule := NewPortugueseFillerWordsRuleWithMinPercent(nil, 0)
	// Sentence-initial capitalised filler must not match (case-sensitive isFillerWord).
	// "Realmente" is not in the set; only "realmente".
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Realmente veio."))))
	sent := languagetool.AnalyzePlain("Ele realmente veio.")
	matches := rule.Match(sent)
	require.Len(t, matches, 1)
	require.Equal(t, "realmente", matchedSurface(sent, matches[0]))
}

func TestPortugueseFillerWordsRule_UserConfigMinPercent(t *testing.T) {
	uc := languagetool.NewUserConfig()
	uc.SetConfigValueByID("FILLER_WORDS_PT", []any{0})
	rule := NewPortugueseFillerWordsRuleWithUserConfig(nil, uc)
	require.Equal(t, 0, rule.GetMinPercent())

	long := "Ele veio de casa com amigos familia jantar gostou noite juntos apos trabalho e realmente saiu cedo ontem."
	sent := languagetool.AnalyzePlain(long)
	matches := rule.Match(sent)
	require.Len(t, matches, 1)
	require.Equal(t, "realmente", matchedSurface(sent, matches[0]))
}

func TestPortugueseFillerWordsRule_MessageKeysNotInvented(t *testing.T) {
	// Without messages, Description must not be invented PT/EN copy.
	rule := NewPortugueseFillerWordsRule(nil)
	require.NotEqual(t, "Filler words", rule.GetDescription())
	require.NotContains(t, rule.GetDescription(), "enchimento")

	msgs := map[string]string{
		"filler_words_rule_desc": "desc-from-messages",
		"filler_words_rule_msg":  "msg-from-messages",
	}
	rule2 := NewPortugueseFillerWordsRuleWithMinPercent(msgs, 0)
	require.Equal(t, "desc-from-messages", rule2.GetDescription())
	matches := rule2.Match(languagetool.AnalyzePlain("Ele realmente veio."))
	require.NotEmpty(t, matches)
	require.Equal(t, "msg-from-messages", matches[0].GetMessage())
}
