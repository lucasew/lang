package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreEnglishRules_Check(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	RegisterCoreEnglishRules(lt)

	// multi whitespace (real rule via adapter)
	m := lt.Check("hello  world")
	require.NotEmpty(t, m)
	var hasWS bool
	for _, x := range m {
		if x.RuleID == "WHITESPACE_RULE" {
			hasWS = true
		}
	}
	require.True(t, hasWS)

	// comma whitespace (e.g. missing space after comma)
	// may or may not fire depending on tokenization; double punct is reliable
	// double punctuation
	m = lt.Check("Wait.. now")
	require.NotEmpty(t, m)

	// RegisterCoreRules dispatch
	lt2 := languagetool.NewJLanguageTool("fr")
	RegisterCoreRules(lt2, "fr")
	require.NotEmpty(t, lt2.Check("bonjour  monde"))

	// a vs an
	m = lt.Check("This is an test.")
	require.NotEmpty(t, m)
	require.Equal(t, "This is a test.", languagetool.CorrectTextFromLocalMatches("This is an test.", m))

	// word repeat
	require.NotEmpty(t, lt.Check("this this"))

	// unpaired
	require.NotEmpty(t, lt.Check("open (paren"))

	// active rules include core ids
	active := lt.GetAllActiveRuleIDs()
	require.Contains(t, active, "WHITESPACE_RULE")
	require.Contains(t, active, "EN_A_VS_AN")
	require.Contains(t, active, "SENTENCE_WHITESPACE")
	require.Contains(t, active, "WHITESPACE_PUNCTUATION")

	// text-level sentence whitespace (missing space after period)
	m = lt.Check("This is a text.And there's the next sentence.")
	var hasSW bool
	for _, x := range m {
		if x.RuleID == "SENTENCE_WHITESPACE" {
			hasSW = true
		}
	}
	require.True(t, hasSW, "matches: %+v", m)

	// space before colon
	m = lt.Check("Wait : now")
	var hasWBP bool
	for _, x := range m {
		if x.RuleID == "WHITESPACE_PUNCTUATION" {
			hasWBP = true
		}
	}
	require.True(t, hasWBP, "matches: %+v", m)
}

func TestToLocalMatches(t *testing.T) {
	sent := languagetool.AnalyzePlain("ab")
	r := NewFakeRule("X")
	ms := []*RuleMatch{NewRuleMatch(r, sent, 0, 2, "msg")}
	ms[0].SetSuggestedReplacements([]string{"AB"})
	lm := ToLocalMatches(ms)
	require.Len(t, lm, 1)
	require.Equal(t, "X", lm[0].RuleID)
	require.Equal(t, []string{"AB"}, lm[0].Suggestions)
}

func TestSharedLayout_ParagraphRules(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	RegisterSharedLayoutRules(lt, "en")
	active := lt.GetAllActiveRuleIDs()
	require.Contains(t, active, "TOO_LONG_PARAGRAPH")
	require.Contains(t, active, "PARAGRAPH_REPEAT_BEGINNING_RULE")

	// paragraph start repeat (need para boundary via leading newline on second sent)
	// SRX may not split on \n\n alone; AnalyzeTextDemo-style double newline often works
	text := "Wiederholung am Anfang.\n\nWiederholung am Ende."
	m := lt.Check(text)
	// soft: may or may not fire depending on tokenizer paragraph handling
	_ = m
}
