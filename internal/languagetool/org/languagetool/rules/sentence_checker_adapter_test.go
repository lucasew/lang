package rules_test

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/fr"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreEnglishRules_Check(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	// Full English pack (wires PreferredAvsAnChecker via en init).
	en.RegisterCoreEnglishLanguageRules(lt)

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

	// French pack: Java getRelevantRules (no invent WordRepeat)
	lt2 := languagetool.NewJLanguageTool("fr")
	fr.RegisterCoreFrenchRules(lt2)
	require.NotEmpty(t, lt2.Check("bonjour  monde"))
	for _, m := range lt2.Check("test test") {
		require.NotContains(t, m.RuleID, "WORD_REPEAT")
	}
	// Unknown codes: no invent SharedLayout
	lt3 := languagetool.NewJLanguageTool("xx")
	rules.RegisterCoreRules(lt3, "xx")
	require.Empty(t, lt3.GetAllRegisteredRuleIDs())

	// a vs an (faithful AvsAnRule + DT inject)
	m = lt.Check("This is an test.")
	require.NotEmpty(t, m)
	require.Equal(t, "This is a test.", languagetool.CorrectTextFromLocalMatches("This is an test.", m))

	// word repeat (faithful WordRepeatRule)
	require.NotEmpty(t, lt.Check("this this"))

	// unpaired
	require.NotEmpty(t, lt.Check("open (paren"))

	// active rules include core ids
	active := lt.GetAllActiveRuleIDs()
	require.Contains(t, active, "WHITESPACE_RULE")
	require.Contains(t, active, "EN_A_VS_AN")
}
