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

	// double punctuation
	m = lt.Check("Wait.. now")
	require.NotEmpty(t, m)

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
