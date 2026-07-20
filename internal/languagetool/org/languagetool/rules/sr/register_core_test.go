package sr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

// Upstream ekavian replace-grammar.txt: еуро=евро (same pair as twin Java test).
func TestRegisterCoreSerbianRules_ReplaceGrammar(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sr")
	RegisterCoreSerbianRules(lt)
	m := lt.Check("То кошта један еуро.")
	found := false
	for _, x := range m {
		if x.RuleID == "SR_EKAVIAN_SIMPLE_GRAMMAR_REPLACE_RULE" {
			found = true
			break
		}
	}
	require.True(t, found, "%+v", m)
}

// Java Serbian.getRelevantRules (Ekavian) exact ID set.
func TestRegisterCoreSerbianRules_JavaRelevantOnly_Ekavian(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sr")
	RegisterCoreSerbianRules(lt)
	require.ElementsMatch(t, language.SerbianEkavianRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())
	for _, bad := range []string{
		"EMPTY_LINE", "TOO_LONG_PARAGRAPH", "WHITESPACE_PUNCTUATION",
		"PARAGRAPH_REPEAT_BEGINNING_RULE", "WHITESPACE_PARAGRAPH",
	} {
		require.NotContains(t, lt.GetAllRegisteredRuleIDs(), bad)
	}
}

// Java JekavianSerbian surface for BA.
func TestRegisterCoreSerbianRules_JavaRelevantOnly_Jekavian(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sr-BA")
	RegisterCoreSerbianRules(lt)
	require.ElementsMatch(t, language.SerbianJekavianRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())
}
